package common

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"testing"

	"github.com/hashicorp/packer/packer/tmp"

	"github.com/google/go-cmp/cmp"

	"github.com/hashicorp/packer/helper/multistep"
)

var _ multistep.Step = new(StepDownload)

func toSha1(in string) string {
	b := sha1.Sum([]byte(in))
	return hex.EncodeToString(b[:])
}

func TestStepDownload_Run(t *testing.T) {
	srvr := httptest.NewServer(http.FileServer(http.Dir("test-fixtures")))
	defer srvr.Close()

	cs := map[string]string{
		"another.txt": "7c6e5dd1bacb3b48fdffba2ed096097eb172497d",
	}

	type fields struct {
		Checksum     string
		ChecksumType string
		Description  string
		ResultKey    string
		TargetPath   string
		Url          []string
		Extension    string
	}

	tests := []struct {
		name      string
		fields    fields
		want      multistep.StepAction
		wantFiles []string
	}{
		{"successfull dl - checksum from url",
			fields{Extension: "txt", Url: []string{srvr.URL + "/root/another.txt?checksum=" + cs["another.txt"]}},
			multistep.ActionContinue,
			[]string{
				toSha1(cs["another.txt"]) + ".txt",
				toSha1(cs["another.txt"]) + ".txt.lock", // a lock file is created & deleted on mac
			},
		},
		{"successfull dl - checksum from parameter - no checksum type",
			fields{Extension: "txt", Url: []string{srvr.URL + "/root/another.txt?"}, Checksum: cs["another.txt"]},
			multistep.ActionContinue,
			[]string{
				toSha1(cs["another.txt"]) + ".txt",
				toSha1(cs["another.txt"]) + ".txt.lock",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := createTempDir(t)
			defer os.RemoveAll(dir)
			s := &StepDownload{
				TargetPath:   tt.fields.TargetPath,
				Checksum:     tt.fields.Checksum,
				ChecksumType: tt.fields.ChecksumType,
				ResultKey:    tt.fields.ResultKey,
				Url:          tt.fields.Url,
				Extension:    tt.fields.Extension,
			}
			defer os.Setenv("PACKER_CACHE_DIR", os.Getenv("PACKER_CACHE_DIR"))
			os.Setenv("PACKER_CACHE_DIR", dir)

			if got := s.Run(context.Background(), testState(t)); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StepDownload.Run() = %v, want %v", got, tt.want)
			}
			files := listFiles(t, dir)
			if diff := cmp.Diff(tt.wantFiles, files); diff != "" {
				t.Fatalf("file list differs in %s: %s", dir, diff)
			}
		})
	}
}

func createTempDir(t *testing.T) string {
	dir, err := tmp.Dir("pkr")
	if err != nil {
		t.Fatalf("err: %s", err)
	}
	return dir
}

func listFiles(t *testing.T, dir string) []string {
	fs, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	var files []string
	for _, file := range fs {
		if file.Name() == "." {
			continue
		}
		files = append(files, file.Name())
	}

	return files
}
