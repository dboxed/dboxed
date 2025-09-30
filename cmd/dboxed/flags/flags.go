package flags

type WebdavProxyFlags struct {
	WebdavProxyListen string `help:"Specify Webdav/S3 proxy listen address" default:"127.0.0.1:0"`
}
