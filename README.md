# edgecontext.py

Documentation: https://reddit-edgecontext.readthedocs.io/en/latest/

Services deep within the backend often need to know information about the
client that originated the request, such as what user is authenticated or what
country they're in. Baseplate services can get this information from the edge
context which is automatically propagated along with calls between services.

This library provides a Thrift specification of an edge context payload and a
corresponding implementation of the [`EdgeContextFactory`] interface from
Baseplate.py.

[`EdgeContextFactory`]: https://baseplate.readthedocs.io/en/latest/api/baseplate/lib/edgecontext.html

## Usage

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
    return request.edgecontext.user.id
```

See [the documentation] for all the available fields.

[the documentation]: https://reddit-edgecontext.readthedocs.io/en/latest/

## Development

A Dockerfile is provided to get a development environment running. To use it,
build the base Docker image:

```console
$ docker build -t edgecontext .
```

And then fire up the environment and use the provided Makefile targets to do
common tasks:

```console
$ docker run -it -v $PWD:/src -w /src edgecontext
$ make fmt
```

The following make targets are provided:

* `fmt`: Apply automatic formatting to the source code.
* `thrift`: Generate code from the Thrift IDL. Run `fmt` after doing this.
* `lint`: Run linters on the code.
* `test`: Run the test suite.
* `docs`: Build the docs. Output can be found in `build/html/`.

The generated Thrift code is committed to the Git repo, so if you change
`reddit_edgecontext/edgecontext.thrift` make sure to run `make thrift fmt` and
commit those changes as well.
