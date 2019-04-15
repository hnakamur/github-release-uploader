package main

import (
	"flag"
	"log"
	"path/filepath"

	"github.com/github/hub/github"
)

func main() {
	release := flag.String("release", "", "release name")
	tag := flag.String("tag", "master", "target commitish")
	assetsGlobPattern := flag.String("assets", "./tmp/*", "asset file glob pattern")
	deleteFlag := flag.Bool("delete", false, "Delete release")
	flag.Parse()

	if *deleteFlag {
		err := deleteRelease(*release)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		err := createRelease(*release, *tag, *assetsGlobPattern)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func createRelease(releaseName, tag, assetsGlobPattern string) error {
	repo, err := github.LocalRepo()
	if err != nil {
		return err
	}
	proj, err := repo.CurrentProject()
	if err != nil {
		return err
	}

	client := github.NewClient(proj.Host)

	log.Printf("creating release: %s, target: %s", releaseName, tag)
	params := &github.Release{
		TagName:         releaseName,
		TargetCommitish: tag,
	}
	release, err := client.CreateRelease(proj, params)
	if err != nil {
		return err
	}

	assetPaths, err := filepath.Glob(assetsGlobPattern)
	if err != nil {
		return err
	}

	for _, assetPath := range assetPaths {
		assetFilename := filepath.Base(assetPath)
		for _, existingAsset := range release.Assets {
			if existingAsset.Name == assetFilename {
				err := client.DeleteReleaseAsset(&existingAsset)
				if err != nil {
					return err
				}
				break
			}
		}
		log.Printf("uploading asset: %s", assetFilename)
		_, err := client.UploadReleaseAsset(release, assetPath, assetFilename)
		if err != nil {
			return err
		}
	}

	return nil
}

func deleteRelease(releaseName string) error {
	repo, err := github.LocalRepo()
	if err != nil {
		return err
	}
	proj, err := repo.CurrentProject()
	if err != nil {
		return err
	}

	client := github.NewClient(proj.Host)

	release, err := client.FetchRelease(proj, releaseName)
	if err != nil {
		return err
	}

	log.Printf("deleting release: %s", releaseName)
	err = client.DeleteRelease(release)
	if err != nil {
		return err
	}

	return nil
}
