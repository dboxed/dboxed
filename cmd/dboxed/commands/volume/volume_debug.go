package commands

type DebugCmd struct {
	WebdavProxy WebdavProxyCmd `cmd:"" help:"Run a webdav proxy"`
}
