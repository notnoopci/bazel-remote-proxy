# bazel-remote-proxy

A remote cache for [Bazel](https://bazel.build) using HTTP/1.1 but that proxies requests to backend storage services (e.g. S3, Google Storage, etc).

## Using bazel-remote-proxy

```
Usage of bazel-remote-proxy:
  -backend string
        uri of backend storage service, e.g. s3://my-bazel-cache/prefix
  -port int
        Port the HTTP server listens to (default 8080)
```

Currently, only S3 is supported as a storage backend.  `bazel-remote-proxy` looks up the AWS credentials and configuration in a similar fashion to the AWS CLI - through `~/.aws/credentials`, env-vars (e.g. `AWS_PROFILE`, `AWS_ACCESS_KEY_ID`, etc) as documented in [CLI docs](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html).
