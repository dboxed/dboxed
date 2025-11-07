package flags

type VolumeMountArgs struct {
	Volume string `help:"Specify volume" required:""`
}

type VolumeServeArgs struct {
	VolumeMountArgs

	BackupInterval string `help:"Specify the backup interval" default:"5m"`
}

type WebdavProxyFlags struct {
	WebdavProxyListen string `help:"Specify Webdav/S3 proxy listen address" default:"127.0.0.1:0"`
}
