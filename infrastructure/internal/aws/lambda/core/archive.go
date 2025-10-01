package lambda_core

import (
	"log"
	"os/exec"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func BuildSourceZip(ctx *pulumi.Context, buildCmd string, workingDir string, outputPath string) (pulumi.AssetOrArchiveInput, error) {
	if buildCmd != "" {
		cmd := exec.Command("sh", "-c", buildCmd)
		cmd.Dir = workingDir
		out, err := cmd.CombinedOutput()
		if err != nil {
			log.Println("Build Lambda Error", "err: ", err)
			return nil, err
		}
		log.Println("build command success: " + string(out))
	}
	return pulumi.NewFileArchive(outputPath), nil
}
