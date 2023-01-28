package apt

import (
	"github.com/codecomet-io/installers/sdk/utils"
)

/**
 * @version mark-I
 * Apt configuration object meant to serialize into apt.conf file via toString()
 */

type Acquire struct {
	utils.SerializableConfig

	UserAgent  string `codecomet:"Acquire::http::User-Agent,omitempty"`
	HTTPProxy  string `codecomet:"Acquire::http::Proxy,omitempty"`
	HTTPSProxy string `codecomet:"Acquire::https::Proxy,omitempty"`

	// Expert section
	// default to false - only turn true if using old snapshots that do not get resigned (DISCOURAGED)
	DoNotCheckIfExpired bool `codecomet:"Acquire::Check-Valid-Until,swap"`
	// default to false - enabling this will reduce disk space for lists, at the (likely) cost of extra processing
	GzipIndexes bool `codecomet:"Acquire::GzipIndexes"`
	// defaults to "none" - and will prevent installation of language files, saving space.
	Languages string `codecomet:"Acquire::Languages"`
	// XXX danger
	// DoNotCheckDate bool
}

/*
func (cnf *Acquire) toString() string {
	com := []string{}
	val := reflect.ValueOf(*cnf)
	typeOf := val.Type()

	for i := 0; i < val.NumField(); i++ {
		serial := typeOf.Field(i).Tag.Get("codecomet")
		if serial == "" {
			serial = strings.ToLower(typeOf.Field(i).Name)
		}
		qual := strings.Split(serial, ",")
		serial = qual[0]
		if typeOf.Field(i).Type.Kind() == reflect.Bool {
			swap := false
			swap = (len(qual) > 1 && qual[1] == "swap")
			value := val.Field(i).Bool()
			if swap {
				value = !value
			}
			com = append(com, fmt.Sprintf(`%s "%t";`, serial, value))
		} else if typeOf.Field(i).Type.Kind() == reflect.String {
			omitempty := false
			omitempty = (len(qual) > 1 && qual[1] == "omitempty")
			value := val.Field(i).String()
			if !omitempty || value != "" {
				com = append(com, fmt.Sprintf(`%s "%s";`, serial, value))
			}
		}
	}
	return strings.Join(com, "\n")
}

/*
func (cnf *Acquire) toString() string {
	content := []string{}

	// UserAgent if need be
	if cnf.UserAgent != "" {
		content = append(content, fmt.Sprintf(`Acquire::http::User-Agent %q;`, cnf.UserAgent))
	}

	// Proxy
	if cnf.HTTPProxy != "" {
		content = append(content, fmt.Sprintf(`Acquire::http::Proxy %q;`, cnf.HTTPProxy))
	}
	if cnf.HTTPSProxy != "" {
		content = append(content, fmt.Sprintf(`Acquire::https::Proxy %q;`, cnf.HTTPSProxy))
	}

	// Useful if using historical snapshots, where packages are not resigned - but by default, using this is a security problem...
	if cnf.DoNotCheckIfExpired == true {
		content = append(content, `Acquire::Check-Valid-Until "no";`)
	} else {
		content = append(content, `Acquire::Check-Valid-Until "yes";`)
	}

	// Used by Docker normally to save space, which is not a concern here
	content = append(content, fmt.Sprintf(`Acquire::GzipIndexes "%t";`, cnf.GzipIndexes))

	// Usually do not need languages, they are just a waste of space
	ln := cnf.Languages
	if ln == "" {
		ln = "none"
	}
	content = append(content, fmt.Sprintf(`Acquire::Languages %q;`, ln))

	return strings.Join(content, "\n")
}
*/

type APT struct {
	utils.SerializableConfig
	// Expert section
	/*
		defaults to "false" - turn true to PREVENT autoremove from doing its job...
		This should be a no-op if InstallRecommends is left to false as well...

		By default, APT will actually _keep_ packages installed via Recommends or
		# Depends if another package Suggests them, even and including if the package
		# that originally caused them to be installed is removed.  Setting this to
		# "false" ensures that APT is appropriately aggressive about removing the
		# packages it added
	*/
	AutoRemoveSuggestsImportant bool `codecomet:"APT::AutoRemove::SuggestsImportant"`

	// Enable this to install recommended packages
	InstallRecommends bool `codecomet:"APT::Install-Recommends"`

	// Should assume yes by default. Given this is not interactive, setting this to false will likely just prevent things from working...
	AssumeYes bool `codecomet:"APT::Get::Assume-Yes"`
}

/*
func (cnf *APT) toString() string {
	content := []string{}

	content = append(content, fmt.Sprintf(`APT::AutoRemove::SuggestsImportant "%t";`, cnf.AutoRemoveSuggestsImportant))
	content = append(content, fmt.Sprintf(`APT::Install-Recommends "%t";`, cnf.InstallRecommends))
	content = append(content, fmt.Sprintf(`APT::Get::Assume-Yes "%t";`, cnf.AssumeYes))

	return strings.Join(content, "\n")
}
*/

type Config struct {
	utils.SerializableConfig

	NetRC       string // `codecomet:"XXX"` // "Dir::Etc::netrc \"\(input.netRC)\";"
	Authority   string // `codecomet:"XXX"` // "Acquire::https::CAInfo \"\(input.authority)\";"
	Certificate string // `codecomet:"XXX"` // "Acquire::https::SSLCert \"\(input.certificate)\";"
	Key         string // `codecomet:"XXX"` // "Acquire::https::SSLKey \"\(input.key)\";"
	// if input.sources != _|_ { "Dir::Etc::SourceList \"\(input.sources)\";"},
	/*
	   // SslForceVersion
	   // FIXME in case we want away from apt-key, we would have to: cat "$SECRET_APT_SOURCES" | sed -E "s,^deb ,^deb [signed-by=/run/secrets/SECRET_GPG] ," > /tmp/sources.list
	   if input.trusted != _|_ { "Dir::Etc::Trusted \"\(input.trusted)\";" },
	   // NOTE because of https://bugs.debian.org/cgi-bin/bugreport.cgi?bug=990555
	   if input.authority != _|_ { "Acquire::http::CAInfo \"\(input.authority)\";" },
	   if input.certificate != _|_ { "Acquire::http::SSLCert \"\(input.certificate)\";" },
	   if input.key != _|_ { "Acquire::http::SSLKey \"\(input.key)\";" },
	*/

	Acquire *Acquire // `codecomet:"recurse"`
	APT     *APT     // `codecomet:"recurse"`
	// XXX additional files
	// Sources string
	// Trusted string

}

func NewConfig() *Config {
	cnf := &Config{
		APT: &APT{
			AutoRemoveSuggestsImportant: false,
			InstallRecommends:           false,
			AssumeYes:                   true,
		},
		Acquire: &Acquire{
			UserAgent:           "codecomet/mark-I",
			HTTPProxy:           "",
			HTTPSProxy:          "",
			DoNotCheckIfExpired: true,
			// Seems like gzipping is impacting the performance with emulation
			GzipIndexes: false,
			Languages:   "",
		},
	}
	cnf.SerializableConfig.KeyValueSeparator = " "
	cnf.SerializableConfig.QuotedValues = true
	cnf.SerializableConfig.EndOfLine = ";\n"
	cnf.APT.SerializableConfig.KeyValueSeparator = " "
	cnf.APT.SerializableConfig.QuotedValues = true
	cnf.APT.SerializableConfig.EndOfLine = ";\n"
	cnf.Acquire.SerializableConfig.KeyValueSeparator = " "
	cnf.Acquire.SerializableConfig.QuotedValues = true
	cnf.Acquire.SerializableConfig.EndOfLine = ";\n"
	return cnf
}

func (cnf *Config) toString() string {
	a := utils.SerializableConfigToString(*cnf.APT)
	b := utils.SerializableConfigToString(*cnf.Acquire)
	c := utils.SerializableConfigToString(*cnf)
	/*
		content := []string{}

		// From 70debconf - see https://manpages.ubuntu.com/manpages/trusty/man8/dpkg-preconfigure.8.html
		// content = append(content, `DPkg::Pre-Install-Pkgs {"/usr/sbin/dpkg-preconfigure --apt || true";};`)

		if cnf.Acquire != nil {
			content = append(content, cnf.Acquire.toString())
		}
		if cnf.APT != nil {
			content = append(content, cnf.APT.toString())
		}

		// Return
		return strings.Join(content, "\n")

	*/
	return a + b + c
}

/*
root@5ae7dd0a92fd:/# apt-config dump
APT "";
APT::Architecture "amd64";
APT::Build-Essential "";
APT::Build-Essential:: "build-essential";
APT::Install-Recommends "1";
APT::Install-Suggests "0";
APT::Sandbox "";
APT::Sandbox::User "_apt";
APT::NeverAutoRemove "";
APT::NeverAutoRemove:: "^firmware-linux.*";
APT::NeverAutoRemove:: "^linux-firmware$";
APT::NeverAutoRemove:: "^linux-image-[a-z0-9]*$";
APT::NeverAutoRemove:: "^linux-image-[a-z0-9]*-[a-z0-9]*$";
APT::VersionedKernelPackages "";
APT::VersionedKernelPackages:: "linux-.*";
APT::VersionedKernelPackages:: "kfreebsd-.*";
APT::VersionedKernelPackages:: "gnumach-.*";
APT::VersionedKernelPackages:: ".*-modules";
APT::VersionedKernelPackages:: ".*-kernel";
APT::Never-MarkAuto-Sections "";
APT::Never-MarkAuto-Sections:: "metapackages";
APT::Never-MarkAuto-Sections:: "contrib/metapackages";
APT::Never-MarkAuto-Sections:: "non-free/metapackages";
APT::Never-MarkAuto-Sections:: "restricted/metapackages";
APT::Never-MarkAuto-Sections:: "universe/metapackages";
APT::Never-MarkAuto-Sections:: "multiverse/metapackages";
APT::Move-Autobit-Sections "";
APT::Move-Autobit-Sections:: "oldlibs";
APT::Move-Autobit-Sections:: "contrib/oldlibs";
APT::Move-Autobit-Sections:: "non-free/oldlibs";
APT::Move-Autobit-Sections:: "restricted/oldlibs";
APT::Move-Autobit-Sections:: "universe/oldlibs";
APT::Move-Autobit-Sections:: "multiverse/oldlibs";
APT::AutoRemove "";
APT::AutoRemove::SuggestsImportant "false";
APT::Update "";
APT::Update::Post-Invoke "";
APT::Update::Post-Invoke:: "rm -f /var/cache/apt/archives/*.deb /var/cache/apt/archives/partial/*.deb /var/cache/apt/*.bin || true";
APT::Architectures "";
APT::Architectures:: "amd64";
APT::Compressor "";
APT::Compressor::. "";
APT::Compressor::.::Name ".";
APT::Compressor::.::Extension "";
APT::Compressor::.::Binary "";
APT::Compressor::.::Cost "0";
APT::Compressor::zstd "";
APT::Compressor::zstd::Name "zstd";
APT::Compressor::zstd::Extension ".zst";
APT::Compressor::zstd::Binary "false";
APT::Compressor::zstd::Cost "60";
APT::Compressor::lz4 "";
APT::Compressor::lz4::Name "lz4";
APT::Compressor::lz4::Extension ".lz4";
APT::Compressor::lz4::Binary "false";
APT::Compressor::lz4::Cost "50";
APT::Compressor::gzip "";
APT::Compressor::gzip::Name "gzip";
APT::Compressor::gzip::Extension ".gz";
APT::Compressor::gzip::Binary "gzip";
APT::Compressor::gzip::Cost "100";
APT::Compressor::gzip::CompressArg "";
APT::Compressor::gzip::CompressArg:: "-6n";
APT::Compressor::gzip::UncompressArg "";
APT::Compressor::gzip::UncompressArg:: "-d";
APT::Compressor::xz "";
APT::Compressor::xz::Name "xz";
APT::Compressor::xz::Extension ".xz";
APT::Compressor::xz::Binary "false";
APT::Compressor::xz::Cost "200";
APT::Compressor::bzip2 "";
APT::Compressor::bzip2::Name "bzip2";
APT::Compressor::bzip2::Extension ".bz2";
APT::Compressor::bzip2::Binary "false";
APT::Compressor::bzip2::Cost "300";
APT::Compressor::lzma "";
APT::Compressor::lzma::Name "lzma";
APT::Compressor::lzma::Extension ".lzma";
APT::Compressor::lzma::Binary "false";
APT::Compressor::lzma::Cost "400";
APT::Compressor::lzma::CompressArg "";
APT::Compressor::lzma::CompressArg:: "--suffix=";
APT::Compressor::lzma::CompressArg:: "-6";
APT::Compressor::lzma::UncompressArg "";
APT::Compressor::lzma::UncompressArg:: "--suffix=";
APT::Compressor::lzma::UncompressArg:: "-d";
Dir "/";
Dir::State "var/lib/apt";
Dir::State::lists "lists/";
Dir::State::cdroms "cdroms.list";
Dir::State::extended_states "extended_states";
Dir::State::status "/var/lib/dpkg/status";
Dir::Cache "var/cache/apt";
Dir::Cache::archives "archives/";
Dir::Cache::srcpkgcache "";
Dir::Cache::pkgcache "";
Dir::Etc "etc/apt";
Dir::Etc::sourcelist "sources.list";
Dir::Etc::sourceparts "sources.list.d";
Dir::Etc::main "apt.conf";
Dir::Etc::netrc "auth.conf";
Dir::Etc::netrcparts "auth.conf.d";
Dir::Etc::parts "apt.conf.d";
Dir::Etc::preferences "preferences";
Dir::Etc::preferencesparts "preferences.d";
Dir::Etc::trusted "trusted.gpg";
Dir::Etc::trustedparts "trusted.gpg.d";
Dir::Bin "";
Dir::Bin::methods "/usr/lib/apt/methods";
Dir::Bin::solvers "";
Dir::Bin::solvers:: "/usr/lib/apt/solvers";
Dir::Bin::planners "";
Dir::Bin::planners:: "/usr/lib/apt/planners";
Dir::Bin::dpkg "/usr/bin/dpkg";
Dir::Bin::gzip "/bin/gzip";
Dir::Bin::bzip2 "/bin/bzip2";
Dir::Bin::xz "/usr/bin/xz";
Dir::Bin::lz4 "/usr/bin/lz4";
Dir::Bin::zstd "/usr/bin/zstd";
Dir::Bin::lzma "/usr/bin/lzma";
Dir::Media "";
Dir::Media::MountPath "/media/apt";
Dir::Log "var/log/apt";
Dir::Log::Terminal "term.log";
Dir::Log::History "history.log";
Dir::Log::Planner "eipp.log.xz";
Dir::Ignore-Files-Silently "";
Dir::Ignore-Files-Silently:: "~$";
Dir::Ignore-Files-Silently:: "\.disabled$";
Dir::Ignore-Files-Silently:: "\.bak$";
Dir::Ignore-Files-Silently:: "\.dpkg-[a-z]+$";
Dir::Ignore-Files-Silently:: "\.ucf-[a-z]+$";
Dir::Ignore-Files-Silently:: "\.save$";
Dir::Ignore-Files-Silently:: "\.orig$";
Dir::Ignore-Files-Silently:: "\.distUpgrade$";
Acquire "";
Acquire::AllowInsecureRepositories "0";
Acquire::AllowWeakRepositories "0";
Acquire::AllowDowngradeToInsecureRepositories "0";
Acquire::cdrom "";
Acquire::cdrom::mount "/media/cdrom/";
Acquire::IndexTargets "";
Acquire::IndexTargets::deb "";
Acquire::IndexTargets::deb::Packages "";
Acquire::IndexTargets::deb::Packages::MetaKey "$(COMPONENT)/binary-$(ARCHITECTURE)/Packages";
Acquire::IndexTargets::deb::Packages::flatMetaKey "Packages";
Acquire::IndexTargets::deb::Packages::ShortDescription "Packages";
Acquire::IndexTargets::deb::Packages::Description "$(RELEASE)/$(COMPONENT) $(ARCHITECTURE) Packages";
Acquire::IndexTargets::deb::Packages::flatDescription "$(RELEASE) Packages";
Acquire::IndexTargets::deb::Packages::Optional "0";
Acquire::IndexTargets::deb::Translations "";
Acquire::IndexTargets::deb::Translations::MetaKey "$(COMPONENT)/i18n/Translation-$(LANGUAGE)";
Acquire::IndexTargets::deb::Translations::flatMetaKey "$(LANGUAGE)";
Acquire::IndexTargets::deb::Translations::ShortDescription "Translation-$(LANGUAGE)";
Acquire::IndexTargets::deb::Translations::Description "$(RELEASE)/$(COMPONENT) Translation-$(LANGUAGE)";
Acquire::IndexTargets::deb::Translations::flatDescription "$(RELEASE) Translation-$(LANGUAGE)";
Acquire::IndexTargets::deb-src "";
Acquire::IndexTargets::deb-src::Sources "";
Acquire::IndexTargets::deb-src::Sources::MetaKey "$(COMPONENT)/source/Sources";
Acquire::IndexTargets::deb-src::Sources::flatMetaKey "Sources";
Acquire::IndexTargets::deb-src::Sources::ShortDescription "Sources";
Acquire::IndexTargets::deb-src::Sources::Description "$(RELEASE)/$(COMPONENT) Sources";
Acquire::IndexTargets::deb-src::Sources::flatDescription "$(RELEASE) Sources";
Acquire::IndexTargets::deb-src::Sources::Optional "0";
Acquire::Changelogs "";
Acquire::Changelogs::URI "";
Acquire::Changelogs::URI::Origin "";
Acquire::Changelogs::URI::Origin::Debian "https://metadata.ftp-master.debian.org/changelogs/@CHANGEPATH@_changelog";
Acquire::Changelogs::URI::Origin::Ubuntu "https://changelogs.ubuntu.com/changelogs/pool/@CHANGEPATH@/changelog";
Acquire::Changelogs::AlwaysOnline "";
Acquire::Changelogs::AlwaysOnline::Origin "";
Acquire::Changelogs::AlwaysOnline::Origin::Ubuntu "1";
Acquire::GzipIndexes "true";
Acquire::Languages "";
Acquire::Languages:: "none";
Acquire::CompressionTypes "";
Acquire::CompressionTypes::xz "xz";
Acquire::CompressionTypes::bz2 "bzip2";
Acquire::CompressionTypes::lzma "lzma";
Acquire::CompressionTypes::gz "gzip";
Acquire::CompressionTypes::lz4 "lz4";
Acquire::CompressionTypes::zst "zstd";
DPkg "";
DPkg::Path "/usr/sbin:/usr/bin:/sbin:/bin";
DPkg::Pre-Install-Pkgs "";
DPkg::Pre-Install-Pkgs:: "/usr/sbin/dpkg-preconfigure --apt || true";
DPkg::Post-Invoke "";
DPkg::Post-Invoke:: "rm -f /var/cache/apt/archives/*.deb /var/cache/apt/archives/partial/*.deb /var/cache/apt/*.bin || true";
Binary "apt-config";
Binary::apt "";
Binary::apt::APT "";
Binary::apt::APT::Color "1";
Binary::apt::APT::Cache "";
Binary::apt::APT::Cache::Show "";
Binary::apt::APT::Cache::Show::Version "2";
Binary::apt::APT::Cache::AllVersions "0";
Binary::apt::APT::Cache::ShowVirtuals "1";
Binary::apt::APT::Cache::Search "";
Binary::apt::APT::Cache::Search::Version "2";
Binary::apt::APT::Cache::ShowDependencyType "1";
Binary::apt::APT::Cache::ShowVersion "1";
Binary::apt::APT::Get "";
Binary::apt::APT::Get::Upgrade-Allow-New "1";
Binary::apt::APT::Get::Update "";
Binary::apt::APT::Get::Update::InteractiveReleaseInfoChanges "1";
Binary::apt::APT::Cmd "";
Binary::apt::APT::Cmd::Show-Update-Stats "1";
Binary::apt::APT::Cmd::Pattern-Only "1";
Binary::apt::APT::Keep-Downloaded-Packages "0";
Binary::apt::DPkg "";
Binary::apt::DPkg::Progress-Fancy "1";
Binary::apt::DPkg::Lock "";
Binary::apt::DPkg::Lock::Timeout "-1";
CommandLine "";
CommandLine::AsString "apt-config dump";
*/
