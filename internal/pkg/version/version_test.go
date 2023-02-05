// Package version
//
// @author: xwc1125
package version

import (
	"fmt"
	"os"
	"testing"
)

func init() {
	FilePath = "./logs/"
	GitHash = "90ee022874f20eff18b83410ab897b3de25392a4"
	GitDate = "1657006174"
	Version = "0.3.0"
	VersionMeta = "stable"
	BuildNumber = "6"
	BuildTime = "2022-07-05T15:43:02+0800"
}

func TestVersion(t *testing.T) {
	os.Args = []string{
		"",
		"build",
	}
	// 0.3.0
	// 0.3.0-stable
	// 0.3.0-stable-90ee0228
	// 0.3.0-stable-90ee0228-1657006174
	Build("App: " + "xwc1125")
	fmt.Println(GetVersion())        // 不带meta的版本：0.3.0.6
	fmt.Println(VersionWithMeta())   // 带meta的版本：0.3.0.6-beta
	fmt.Println(ArchiveVersion())    // 带meta和commit的版本：0.3.0.6-beta-90ee0228
	fmt.Println(VersionWithCommit()) // 带meta、commit和commit日期的版本：0.3.0.6-beta-90ee0228-1657006174
}
