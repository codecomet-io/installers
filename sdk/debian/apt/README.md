# Apt, caching, filesystem and concurrency

Most importantly, apt is not safe to be used concurrently.

Apt relies on two key locations:
 * `/var/lib/apt/lists` for the package lists
 * `/var/cache/apt/archives` for the downloaded packages

These locations can be configured respectively with `Dir::State::Lists`
and `Dir::Cache::Archives`.

`Dir::State::Lists` will be wiped out by `apt-get update` if:
 * `sources.list` differs
 * architecture differs

For `Dir::Cache::Archives` it looks like unfortunately apt will try to get 
an exclusive lock file even when it doesn't need one (eg: no-download), which
annoyingly kills the whole idea of widely shared caching for downloaded packages.

Furthermore, separating download and no-download phases comes at a cost (4s).

## On buildkit volumes

* private or locked are the only realistic choices for raw use of buildkit shares,
as apt requires exclusive use of its locations
* obviously locked will queue parallel builds, which is quite bad
* we do workaround this with a mix of out of band locking and selective copying / symlinking

## Take-away

`apt-get update`
 * only manipulates `Dir::State::Lists`
 * needs exclusive access to it
 * needs this to be sharded per distro and per architecture

Other `apt` (upgrade, install) operations:
 * need *read* access to `Dir::State::Lists`, and can do so concurrently (no locking)
 * need write access AND exclusive lock to `Dir::Cache::Archives`

Possilities:
 1. no caching
    * this is the baseline
    * pros:
      * possibly `sequence 100%`: subsequent operations may have a guarantee that `list` files exist,
      * `parallel` update operations is a given
    * cons:
      * either `clutter` or `sequence 0%`: files are committed to the image if we keep the files around, or we pay the cold cost everytime we update if we delete the files 
      * `no-cache`: there is no concept of shared cache accross distinct stages / builds
 2. use `Dir::State::Lists` in a volume
    * pros:
      * `cache`: accross distinct stages and builds (cuts the cost of `update` in half)
      * `no-clutter`
    * cons:
      * `sequence 50%`: no guarantee that update files have been kept and every other operation will have to update again (though likely cached)
      * `no-parallel` update operations
 3. use `Dir::State::Lists` in a tmpFS
    * pros:
      * `parallel`
      * `no-clutter`
    * cons:
      * `no-cache`
      * `sequence 0%`
 4. use `Dir::State::Lists` in a tmpFS backed by a shared volume
    * pros:
      * `parallel`
      * `cache`
      * `no-clutter`
    * cons:
      * `sequence 50%`: no guarantee that update files have been kept and every other operation will have to update again (though likely cached)
 5. use `Dir::State::Lists` in the state, backed by a shared volume
    * pros:
      * `parallel`
      * `cache`
      * `sequence 100%`: no guarantee that update files have been kept and every other operation will have to update again (though likely cached)
    * cons:
      * `clutter`

Approach 1., because of `no-cache` is the problematic one obviously.

Approach 2. (buildkit shares) will create a large number of shares (because of the necessary sharding)
that may well be garbage collected too often. Furthermore, given apt requirements, locking
is unavoidable and will be a serious bottleneck. This scenario likely has only benefits
for environment where nothing is done in parallel, and mostly building the same thing over.

Approach 3. (use a tmpfs) is the best option to minimize any changes to the state, but
also the worse in all other regards.

Approach 4. is strong, but implies that we need to update before any op.

Approach 5. trades of the above need at the cost of clutter in the state.

## Numbers

### Copy over into a tmpfs

| Op                     | Load          | Process | Save            |
|------------------------|---------------|---------|-----------------|
| Cold start, nothing    | 0.25s (no-op) | 11s     | 0.35s           |
| Subsequent run         | 0.28s         | 4.35s   | 0.237s (no-op)  |

Extra cost per run is about 0.6s in all scenarios.

This will save about 6 seconds every time another worker needs to actually (eg: no bk cache)
run `apt-get update`.
This is true as long as the volume is kept.

Though, this approach makes it necessary to `update` everytime we want to `install` something,
since there is no guarantee the shared mount is up-to-date.
The cost is thus 5 seconds per subsequent `install|upgrade` in the same state.

### Copy over into the state

This is a net positive for sequential operations on the same state (as repeat `update` are no longer needed).

The downside being the state will end-up in the final image (17MB for bullseye, uncompressed).

### Summary

| State           | Op                | in state (*) | in tmpfs | in share (**) | in tmpfs with share | in state with share (*) |
|-----------------|-------------------|--------------|----------|---------------|---------------------|-------------------------|
| Current state   | first update      | 9s           | 9s       | 9s            | 9s+0.5s             | 9s+0.5s                 |
| Current state   | subsequent update | 4.5s         | 9s       | 4.5s          | 4.5s+0.5s           | 4.5s+0.5s               |
| Current state   | upgrade / install | 0s           | 9s       | 4.5s          | 4.5s+0.5s           | 0s                      |
| Repeat build    | first update      | 9s           | 9s       | 4.5s          | 4.5s+0.5s           | 0s                      |
| Repeat build    | subsequent update | 4.5s         | 9s       | 4.5s          | 4.5s+0.5s           | 0s                      |
| Repeat build    | upgrade / install | 0s           | 9s       | 4.5s          | 4.5s+0.5s           | 0s                      |
| Different state | first update      | 9s           | 9s       | 4.5s          | 4.5s+0.5s           | 4.5s+0.5s               |
| Different state | subsequent update | 4.5s         | 9s       | 4.5s          | 4.5s+0.5s           | 0s                      |
| Different state | upgrade / install | 0s           | 9s       | 4.5s          | 4.5s+0.5s           | 0s                      |

(*) creates clutter
(**) kills parallelism

A4. is likely the best option if there will only be one operation in the state

A5. is better in case we will have multiple successive operations, at the cost of clutter.

### About packages

Splitting download and no-download operation comes at a cost of about 4s, with the only
upside being faster availability of cached packages to other operations - thus, not worth the cost.

Cache linking-over is also much faster than copying.

Finally, there is no upside in keeping the overlay in the state.

So, in tmpfs with share is definitely the right option for packages, with no download/no-download separation.


## Parting thoughts

Things to know:
* by default, official images do prevent package caching, and are also using settings meant to
  optimize stuff for Dockerfile scenario. These are detrimental in our case, since we DO want to keep things around.
 Thus, default apt configuration is just ignored and overridden entirely.
* it is unclear to what extent apt can share cache folders with different architectures, different versions of the linux distribution,
  or even different distributions, without issues
* given buildkit caching model, there is no guarantee that the cache is in a usable state across successive operations
* an apt-proxy bundled with buildkitd is probably a good alternative to all these shenanigans...

