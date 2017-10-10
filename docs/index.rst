########################
 The ``quantum`` manual
########################

Welcome to the ``quantum`` manual. ``quantum`` is a lightweight and WAN oriented software defined network (SDN), that is designed with security and scalability at its heart. This manual contains a full reference to ``quantum``, and indepth descriptions of its inner workings.

Features
========

The high level features of ``quantum`` are simple:

  * Thrives in a high latency and distributed infrastructure.
  * Gracefully handles and manages network partitions.
  * Transparently provides powerful network level plugins to augment traffic in real time.
  * Provides a single seamless fully authenticated and secured mesh network.
  * Exposes powerful network metrics for all connected peers, down to which network queue transmitted/received a packet.


.. list-table::
   :widths: auto
   :header-rows: 1
   :align: center

   * - Getting Started
     - Administration
     - Development
     - Reference
   * - :doc:`Introduction <introduction>`
     - :doc:`Configuration Options <configuration>`
     - `Godoc <https://godoc.org/github.com/supernomad/quantum/>`_
     - :doc:`Get Help <help>`
   * - :doc:`Install <install>`
     - :doc:`Operation <operation>`
     - :doc:`Contributing <contribute>`
     - :doc:`How it works <how-it-works>`
   * - :doc:`Quick Start <quick-start>`
     - :doc:`Security <security>`
     -
     -

External Resources
==================



Special Thanks
==============

I would like to reach out to a few specific OSS projects that made ``quantum`` possible.

First and foremost I would like to give a special thanks to the different software defined networks that exist, without which I would have never been inspired to make this project. So hats off to all of you.

I would also like to thank all of the authors and collaborators of the following OSS libraries which ``quantum`` relies on to function, and wouldn't exist without:
  * `OpenSSL <https://www.openssl.org/>`_
  * `Etcd Client <https://github.com/coreos/etcd>`_
  * `Semver <https://github.com/coreos/go-semver/semver>`_
  * `Snappy <https://github.com/golang/snappy>`_
  * `Codec <https://github.com/ugorji/go/codec>`_
  * `Netlink <https://github.com/vishvananda/netlink>`_
  * `Netns <https://github.com/vishvananda/netns>`_
  * `Curve25519 <https://golang.org/x/crypto/curve25519>`_
  * `Pbkdf2 <https://golang.org/x/crypto/pbkdf2>`_
  * `Context <https://golang.org/x/net/context>`_
  * `Yaml <https://gopkg.in/yaml.v2>`_

.. toctree::
   :hidden:
   :maxdepth: 2

   introduction.rst
   install.rst
   quick-start.rst
   how-it-works.rst
   configuration.rst
   operation.rst
   security.rst
   network-backends.rst
   plugins.rst
   contribute.rst
   help.rst
   license.rst
