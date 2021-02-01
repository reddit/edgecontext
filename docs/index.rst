``reddit_edgecontext``
----------------------

.. automodule:: reddit_edgecontext

When a user initiates a request, only the service at the "edge" is directly
talking to them. Baseplate provides for data that is automatically propagated
from service to service so that services far from the user can know about the
client too. This library describes the fields of that data.

To set the library up, create an
:py:class:`~reddit_edgecontext.EdgeContextFactory` and pass it to the
Baseplate.py framework integration you're using:

.. code-block:: python

   from baseplate import Baseplate
   from baseplate.lib.secrets import secrets_store_from_config
   from reddit_edgecontext import EdgeContextFactory


   def make_processor(app_config):
       secrets = secrets_store_from_config(app_config, timeout=60)
       edgecontext_factory = EdgeContextFactory(secrets)

       # pass edgecontext_factory to your framework's integration
       # for Thrift: baseplate.frameworks.thrift.baseplateify_processor
       # for Pyramid: baseplate.frameworks.pyramid.BaseplateConfigurator

Once that's done, you can access the data in the edge context by using the
``edge_context`` attribute on the request object:

.. code-block:: python

   def my_view(request):
       return f"Hi {request.edge_context.user.id}!"

See below for all the fields available in the edge context payload.

The edge context payload
========================

.. autoclass:: EdgeContext
   :members:

.. autoclass:: User
   :members:

.. autoclass:: OAuthClient
   :members:

.. autoclass:: Session
   :members:

.. autoclass:: Service
   :members:

.. autoclass:: AuthenticationToken
   :members:


The factory
===========

.. autoclass:: EdgeContextFactory
   :members:

Errors
======
.. autoexception:: NoAuthenticationError
