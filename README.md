# edgecontext

Python Documentation: https://reddit-edgecontext.readthedocs.io/en/latest/

Go Documentation: https://pkg.go.dev/github.com/reddit/edgecontext/lib/go/edgecontext

Services deep within the backend often need to know information about the
client that originated the request, such as what user is authenticated or what
country they're in. Baseplate services can get this information from the edge
context which is automatically propagated along with calls between services.

This library provides a Thrift specification of an edge context payload and a
corresponding implementation of the [`EdgeContextFactory`] interface from
Baseplate.py.

[`EdgeContextFactory`]: https://baseplate.readthedocs.io/en/latest/api/baseplate/lib/edgecontext.html

And an implementation of [`ecinterface`] from Baseplate.go.

[`ecinterface`]: https://pkg.go.dev/github.com/reddit/baseplate.go/ecinterface

## Usage

### Python

Add the `EdgeContextFactory` to application startup:

```python
from baseplate import Baseplate
from baseplate.lib.secrets import secrets_store_from_config
from reddit_edgecontext import EdgeContextFactory


def make_processor(app_config):
    secrets = secrets_store_from_config(app_config, timeout=60)
    edgecontext_factory = EdgeContextFactory(secrets)

    # pass edgecontext_factory to your framework's integration
    # for Thrift: baseplate.frameworks.thrift.baseplateify_processor
    # for Pyramid: baseplate.frameworks.pyramid.BaseplateConfigurator
```

Then read fields while handling requests:

```python
def my_view(request):
    return request.edge_context.user.id
```

See [the documentation] for all the available fields.

[the documentation]: https://reddit-edgecontext.readthedocs.io/en/latest/

### Go

Use [`edgecontext.Factory`] to create an [`ecinterface.Factory`] implementation
that's expected by [`baseplate.New`]:

```go
ctx, bp, err := baseplate.New(context.Background(), baseplate.NewArgs{
  ConfigPath: configPath,
  ServiceCfg: &cfg, // or nil if you don't have additional config to parse
  EdgeContextFactory: edgecontext.Factory(edgecontext.Config{
    Logger: log.ErrorWithSentryWrapper(),
  }),
})
```

When using it, get the [`*EdgeRequestContext`] object out of context:

```go
if ec, ok := edgecontext.GetEdgeContext(ctx); ok {
  user := ec.User()
  loid, ok := user.LoID()
  // Do something with loid
}
```

[`edgecontext.Factory`]: https://pkg.go.dev/github.com/reddit/edgecontext/lib/go/edgecontext#Factory
[`ecinterface.Factory`]: https://pkg.go.dev/github.com/reddit/baseplate.go/ecinterface#Factory
[`baseplate.New`]: https://pkg.go.dev/github.com/reddit/baseplate.go#New
[`*EdgeRequestContext`]: https://pkg.go.dev/github.com/reddit/edgecontext/lib/go/edgecontext#EdgeRequestContext

## Development

A Dockerfile is provided to get a development environment running. To use it,
build the base Docker image:

```console
$ docker build -t edgecontext .
```

And then fire up the environment and use the provided Makefile targets to do
common tasks:

```console
$ docker run -it -v $PWD:/src --user "$(id -u):$(id -g)" -w /src edgecontext
$ make fmt
```

The following make targets are provided:

* `fmt`: Apply automatic formatting to the source code.
* `thrift`: Generate code from the Thrift IDL. Run `fmt` after doing this.
* `lint`: Run linters on the code.
* `test`: Run the test suite.
* `docs`: Build docs.
    * Python output can be found in `lib/py/build/html/`.

The generated Thrift code is committed to the Git repo, so if you change
`edgecontext.thrift` make sure to run `make thrift fmt` and commit those
changes as well.

For Go, we do the same linting checks as Baseplate.go,
so please follow Baseplate.go's [Editor] guide to make sure you are doing the
same linting locally correctly.
Please also follow Baseplate.go's [Style] guide for code style.

[Editor]: https://github.com/reddit/baseplate.go/blob/master/Editor.md
[Style]: https://github.com/reddit/baseplate.go/blob/master/Style.md
