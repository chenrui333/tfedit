package awsv4upgrade

import (
	"testing"

	"github.com/minamijoyo/hcledit/editor"
)

func TestAWSS3BucketFilter(t *testing.T) {
	cases := []struct {
		name string
		src  string
		ok   bool
		want string
	}{
		{
			name: "simple",
			src: `
resource "aws_s3_bucket" "example" {
  bucket = "tfedit-test"
  acl    = "private"

  lifecycle_rule {
    id      = "Keep previous version 30 days, then in Glacier another 60"
    enabled = true

    noncurrent_version_transition {
      days          = 30
      storage_class = "GLACIER"
    }

    noncurrent_version_expiration {
      days = 90
    }
  }

  lifecycle_rule {
    id                                     = "Delete old incomplete multi-part uploads"
    enabled                                = true
    abort_incomplete_multipart_upload_days = 7
  }

  logging {
    target_bucket = "tfedit-test-log"
    target_prefix = "log/"
  }
}
`,
			ok: true,
			// TODO: vertical format
			want: `
resource "aws_s3_bucket" "example" {
  bucket = "tfedit-test"



}

resource "aws_s3_bucket_acl" "example" {
  bucket = aws_s3_bucket.example.id
  acl    = "private"
}

resource "aws_s3_bucket_lifecycle_configuration" "example" {
  bucket = aws_s3_bucket.example.id

  rule {
    id = "Keep previous version 30 days, then in Glacier another 60"

    noncurrent_version_transition {
      storage_class   = "GLACIER"
      noncurrent_days = 30
    }

    noncurrent_version_expiration {
      noncurrent_days = 90
    }
    status = "Enabled"

    filter {
      prefix = ""
    }
  }

  rule {
    id     = "Delete old incomplete multi-part uploads"
    status = "Enabled"

    filter {
      prefix = ""
    }

    abort_incomplete_multipart_upload {
      days_after_initiation = 7
    }
  }
}

resource "aws_s3_bucket_logging" "example" {
  bucket = aws_s3_bucket.example.id

  target_bucket = "tfedit-test-log"
  target_prefix = "log/"
}
`,
		},
		{
			name: "resource type not found",
			src: `
resource "aws_s3_bucket_foo" "example" {
  bucket = "tfedit-test"
}
`,
			ok: true,
			want: `
resource "aws_s3_bucket_foo" "example" {
  bucket = "tfedit-test"
}
`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			filter := &AWSS3BucketFilter{}
			o := editor.NewEditOperator(filter)
			output, err := o.Apply([]byte(tc.src), "test")
			if tc.ok && err != nil {
				t.Fatalf("unexpected err = %s", err)
			}

			got := string(output)
			if !tc.ok && err == nil {
				t.Fatalf("expected to return an error, but no error, outStream: \n%s", got)
			}

			if got != tc.want {
				t.Fatalf("got:\n%s\nwant:\n%s", got, tc.want)
			}
		})
	}
}
