package flags

type VolumeServeArgs struct {
	Volume         string  `help:"Specify volume" required:"" arg:""`
	BackupInterval string  `help:"Specify the backup interval" default:"5m"`
	Box            *string `help:"Specify the box that wants to serve this volume"`
}

type WebdavProxyFlags struct {
	WebdavProxyListen string `help:"Specify Webdav/S3 proxy listen address" default:"127.0.0.1:0"`
}
