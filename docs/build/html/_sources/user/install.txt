.. _install:

Installation
============
This part of the documentation covers the installation of the Go
Elasticsearch Alerts binary.

Download
--------
You can download your preferred variant of the binary from the
`releases page <https://github.com/morningconsult/go-elasticsearch-alerts/releases>`_
of this project's GitHub repository.

Go Get
------
If you have Go installed locally, you can build the binary via
``go get``::

    $ go get github.com/morningconsult/go-elasticsearch-alerts

Once the command finishes, the binary can be found in
``$GOPATH/bin``.

Docker
------
If you do not have Go installed locally, you can still build the
binary if you have Docker installed. Simply clone this repository
and run ``make docker`` to build the binary within a Docker
container and output it to the local directory::

    $ git clone https://github.com/morningconsult/go-elasticsearch-alerts.git
    $ cd go-elasticsearch-alerts
    $ make docker

Once the command finishes, the binary can be found in the ``bin``
directory.

You can also cross-compile the binary using the ``$TARGET_GOOS``
and ``$TARGET_GOARCH`` environment variables. For example, if you
wish to compile the binary for a 64-bit (x86-64) Windows machine,
run the following command::

    $ TARGET_GOOS="windows" TARGET_GOARCH="amd64" make docker
