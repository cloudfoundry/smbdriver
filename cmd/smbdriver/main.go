package main

import (
	cf_http "code.cloudfoundry.org/cfhttp"
	cf_debug_server "code.cloudfoundry.org/debugserver"
	"code.cloudfoundry.org/dockerdriver"
	"code.cloudfoundry.org/dockerdriver/driverhttp"
	"code.cloudfoundry.org/dockerdriver/invoker"
	"code.cloudfoundry.org/goshims/bufioshim"
	"code.cloudfoundry.org/goshims/filepathshim"
	"code.cloudfoundry.org/goshims/ioutilshim"
	"code.cloudfoundry.org/goshims/osshim"
	"code.cloudfoundry.org/goshims/timeshim"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagerflags"
	"code.cloudfoundry.org/smbdriver"
	"code.cloudfoundry.org/smbdriver/driveradmin/driveradminhttp"
	"code.cloudfoundry.org/smbdriver/driveradmin/driveradminlocal"
	"code.cloudfoundry.org/volumedriver"
	"code.cloudfoundry.org/volumedriver/mountchecker"
	"code.cloudfoundry.org/volumedriver/oshelper"
	"encoding/json"
	"flag"
	"github.com/tedsuo/ifrit"
	"github.com/tedsuo/ifrit/grouper"
	"github.com/tedsuo/ifrit/http_server"
	"github.com/tedsuo/ifrit/sigmon"
	"os"
	"path/filepath"
	"strconv"
)

var atPort = flag.Int(
	"listenPort",
	8589,
	"Port to serve volume management functions. Listen address is always 127.0.0.1",
)

var adminPort = flag.Int(
	"adminPort",
	8590,
	"Port to serve process admin functions",
)

var driversPath = flag.String(
	"driversPath",
	"",
	"[REQUIRED] - Path to directory where drivers are installed",
)

var transport = flag.String(
	"transport",
	"tcp",
	"Transport protocol to transmit HTTP over",
)

var mountDir = flag.String(
	"mountDir",
	"/tmp/volumes",
	"Path to directory where fake volumes are created",
)

var requireSSL = flag.Bool(
	"requireSSL",
	false,
	"Whether the fake driver should require ssl-secured communication",
)

var caFile = flag.String(
	"caFile",
	"",
	"(optional) - The certificate authority public key file to use with ssl authentication",
)

var certFile = flag.String(
	"certFile",
	"",
	"(optional) - The public key file to use with ssl authentication",
)

var keyFile = flag.String(
	"keyFile",
	"",
	"(optional) - The private key file to use with ssl authentication",
)
var clientCertFile = flag.String(
	"clientCertFile",
	"",
	"(optional) - The public key file to use with client ssl authentication",
)

var clientKeyFile = flag.String(
	"clientKeyFile",
	"",
	"(optional) - The private key file to use with client ssl authentication",
)

var insecureSkipVerify = flag.Bool(
	"insecureSkipVerify",
	false,
	"Whether SSL communication should skip verification of server IP addresses in the certificate",
)

var mountFlagAllowed = flag.String(
	"mountFlagAllowed",
	"",
	"[REQUIRED] - This is a comma separted list of parameters allowed to be send in extra config. Each of this parameters can be specify by brokers",
)

var mountFlagDefault = flag.String(
	"mountFlagDefault",
	"",
	"(optional) - This is a comma separted list of like params:value. This list specify default value of parameters. If parameters has default value and is not in allowed list, this default value become a forced value who's cannot be override",
)

var uniqueVolumeIds = flag.Bool(
	"uniqueVolumeIds",
	false,
	"Whether the SMB driver should opt-in to unique volumes",
)

const listenAddress = "127.0.0.1"

func main() {
	parseCommandLine()

	var smbDriverServer ifrit.Runner

	logger, logSink := newLogger()
	logger.Info("start")
	defer logger.Info("end")

	configMask, err := smbdriver.NewSmbVolumeMountMask(*mountFlagAllowed, *mountFlagDefault)
	exitOnFailure(logger, err)

	mounter := smbdriver.NewSmbMounter(
		invoker.NewProcessGroupInvoker(),
		&osshim.OsShim{},
		&ioutilshim.IoutilShim{},
		configMask,
	)

	client := volumedriver.NewVolumeDriver(
		logger,
		&osshim.OsShim{},
		&filepathshim.FilepathShim{},
		&ioutilshim.IoutilShim{},
		&timeshim.TimeShim{},
		mountchecker.NewChecker(&bufioshim.BufioShim{}, &osshim.OsShim{}),
		*mountDir,
		mounter,
		oshelper.NewOsHelper(),
	)

	if *transport == "tcp" {
		smbDriverServer = createSmbDriverServer(logger, client, *atPort, *driversPath, false, false)
	} else if *transport == "tcp-json" {
		smbDriverServer = createSmbDriverServer(logger, client, *atPort, *driversPath, true, *uniqueVolumeIds)
	} else {
		smbDriverServer = createSmbDriverUnixServer(logger, client, *atPort)
	}

	servers := grouper.Members{
		{Name: "smbdriver-server", Runner: smbDriverServer},
	}

	if dbgAddr := cf_debug_server.DebugAddress(flag.CommandLine); dbgAddr != "" {
		servers = append(grouper.Members{
			{Name: "debug-server", Runner: cf_debug_server.Runner(dbgAddr, logSink)},
		}, servers...)
	}

	adminClient := driveradminlocal.NewDriverAdminLocal()
	adminHandler, _ := driveradminhttp.NewHandler(logger, adminClient)
	adminAddress := listenAddress + ":" + strconv.Itoa(*adminPort)
	adminServer := http_server.New(adminAddress, adminHandler)

	servers = append(grouper.Members{
		{Name: "driveradmin", Runner: adminServer},
	}, servers...)

	process := ifrit.Invoke(processRunnerFor(servers))
	logger.Info("started")

	adminClient.SetServerProc(process)
	adminClient.RegisterDrainable(client)

	untilTerminated(logger, process)
}

func exitOnFailure(logger lager.Logger, err error) {
	if err != nil {
		logger.Fatal("fatal-err-aborting", err)
	}
}

func untilTerminated(logger lager.Logger, process ifrit.Process) {
	err := <-process.Wait()
	exitOnFailure(logger, err)
}

func processRunnerFor(servers grouper.Members) ifrit.Runner {
	return sigmon.New(grouper.NewOrdered(os.Interrupt, servers))
}

func createSmbDriverServer(logger lager.Logger, client dockerdriver.Driver, atPort int, driversPath string, jsonSpec bool, uniqueVolumeIds bool) ifrit.Runner {
	atAddress := listenAddress + ":" + strconv.Itoa(atPort)
	advertisedUrl := "http://" + atAddress
	logger.Info("writing-spec-file", lager.Data{"location": driversPath, "name": "smbdriver", "address": advertisedUrl, "unique-volume-ids": uniqueVolumeIds})
	if jsonSpec {
		driverJsonSpec := dockerdriver.DriverSpec{Name: "smbdriver", Address: advertisedUrl, UniqueVolumeIds: uniqueVolumeIds}

		if *requireSSL {
			absCaFile, err := filepath.Abs(*caFile)
			exitOnFailure(logger, err)
			absClientCertFile, err := filepath.Abs(*clientCertFile)
			exitOnFailure(logger, err)
			absClientKeyFile, err := filepath.Abs(*clientKeyFile)
			exitOnFailure(logger, err)
			driverJsonSpec.TLSConfig = &dockerdriver.TLSConfig{InsecureSkipVerify: *insecureSkipVerify, CAFile: absCaFile, CertFile: absClientCertFile, KeyFile: absClientKeyFile}
			driverJsonSpec.Address = "https://" + atAddress
		}

		jsonBytes, err := json.Marshal(driverJsonSpec)

		exitOnFailure(logger, err)
		err = dockerdriver.WriteDriverSpec(logger, driversPath, "smbdriver", "json", jsonBytes)
		exitOnFailure(logger, err)
	} else {
		err := dockerdriver.WriteDriverSpec(logger, driversPath, "smbdriver", "spec", []byte(advertisedUrl))
		exitOnFailure(logger, err)
	}

	handler, err := driverhttp.NewHandler(logger, client)
	exitOnFailure(logger, err)

	var server ifrit.Runner
	if *requireSSL {
		tlsConfig, err := cf_http.NewTLSConfig(*certFile, *keyFile, *caFile)
		if err != nil {
			logger.Fatal("tls-configuration-failed", err)
		}
		server = http_server.NewTLSServer(atAddress, handler, tlsConfig)
	} else {
		server = http_server.New(atAddress, handler)
	}

	return server
}

func createSmbDriverUnixServer(logger lager.Logger, client dockerdriver.Driver, atPort int) ifrit.Runner {
	atAddress := listenAddress + ":" + strconv.Itoa(atPort)
	handler, err := driverhttp.NewHandler(logger, client)
	exitOnFailure(logger, err)
	return http_server.NewUnixServer(atAddress, handler)
}

func newLogger() (lager.Logger, *lager.ReconfigurableSink) {
	lagerConfig := lagerflags.ConfigFromFlags()
	lagerConfig.RedactSecrets = true
	lagerConfig.RedactPatterns = SmbRedactValuePatterns()

	return lagerflags.NewFromConfig("smb-driver-server", lagerConfig)
}

func parseCommandLine() {
	lagerflags.AddFlags(flag.CommandLine)
	cf_debug_server.AddFlags(flag.CommandLine)
	flag.Parse()
}



func SmbRedactValuePatterns() []string {
	nfsPasswordPattern := `.*password.*`
	awsAccessKeyIDPattern := `AKIA[A-Z0-9]{16}`
	awsSecretAccessKeyPattern := `KEY["']?\s*(?::|=>|=)\s*["']?[A-Z0-9/\+=]{40}["']?`
	cryptMD5Pattern := `\$1\$[A-Z0-9./]{1,16}\$[A-Z0-9./]{22}`
	cryptSHA256Pattern := `\$5\$[A-Z0-9./]{1,16}\$[A-Z0-9./]{43}`
	cryptSHA512Pattern := `\$6\$[A-Z0-9./]{1,16}\$[A-Z0-9./]{86}`
	privateKeyHeaderPattern := `-----BEGIN(.*)PRIVATE KEY-----`

	return []string{nfsPasswordPattern, awsAccessKeyIDPattern, awsSecretAccessKeyPattern, cryptMD5Pattern, cryptSHA256Pattern, cryptSHA512Pattern, privateKeyHeaderPattern}
}
