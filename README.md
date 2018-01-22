# bazel-remote-proxy

A remote cache for [Bazel](https://bazel.build) using HTTP/1.1 but that proxies requests to backend storage services (e.g. S3, Google Storage, etc).

## Using bazel-remote-proxy

```
Usage of bazel-remote-proxy:
  -backend string
        uri of backend storage service, e.g. s3://my-bazel-cache/prefix
  -bind string
        address and port to bind to (default "127.0.0.1:7643")
```

There are few supported storage backends:

### S3

S3 bucket can be used as a centralized caching storage for bazel, by passing S3 path as a `backend` argument, e.g. `s3://bucket-name/prefix`

`bazel-remote-proxy` looks up the AWS credentials and configuration in a similar fashion to the AWS CLI - through `~/.aws/credentials`, env-vars (e.g. `AWS_PROFILE`, `AWS_ACCESS_KEY_ID`, etc) as documented in [CLI docs](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-getting-started.html).

### CircleCI

When running inside CircleCI 2.0 build environment, you can use CircleCI caching storage - eliminating the need to manage one's own S3 credentials and bucket.

A sample configuration would be:

```yaml
    # step 1. install the binary
    - run:
        name: install bazel-remote-proxy
        command: |
          # if go is already installed
          # go install github.com/notnoopci/bazel-remote-proxy

          # otherwise download latest artifact
          DOWNLOAD_URL="$(curl -sSL \
             https://circleci.com/api/v1.1/project/github/notnoopci/bazel-remote-proxy/latest/artifacts?branch=master \
             | grep -o -e "https://[^\"]*/bazel-remote-proxy-$(uname -s)_$(uname -m)" \
          )"

          curl -o ~/bazel-remote-proxy "$DOWNLOAD_URL"
          chmod +x ~/bazel-remote-proxy

    # step 2. start the proxy
    - run:
        name: setup bazel remote proxy
        command: ~/bazel-remote-proxy -backend circleci://
        background: true

    # step 3. configure bazel to use cache:
    - run:
        name: build
        command: |
          bazel \
            --host_jvm_args=-Dbazel.DigestFunction=sha256 \
            build \
            --spawn_strategy=remote \
            --strategy=Javac=remote \
            --genrule_strategy=remote \
            --remote_rest_cache=http://localhost:7654 \
            //foo:target
```
