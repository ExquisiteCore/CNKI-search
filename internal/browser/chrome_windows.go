//go:build windows

package browser

func osChromeCandidates() []string {
	return []string{
		expandEnv(`${ProgramFiles}\Google\Chrome\Application\chrome.exe`),
		expandEnv(`${ProgramFiles(x86)}\Google\Chrome\Application\chrome.exe`),
		expandEnv(`${LocalAppData}\Google\Chrome\Application\chrome.exe`),
		expandEnv(`${ProgramFiles}\Microsoft\Edge\Application\msedge.exe`),
		expandEnv(`${ProgramFiles(x86)}\Microsoft\Edge\Application\msedge.exe`),
	}
}
