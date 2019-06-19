.. Go Elasticsearch Alerts documentation master file, created by
   sphinx-quickstart on Fri Dec  7 10:44:19 2018.
   You can adapt this file completely to your liking, but it should at least
   contain the root `toctree` directive.

Go Elasticsearch Alerts
=======================
Release |release|. (:ref:`Installation <install>`)

.. image:: https://ci.morningconsultintelligence.com/api/v1/teams/oss/pipelines/go-elasticsearch-alerts/jobs/build-release/badge
    :target: https://ci.morningconsultintelligence.com/teams/oss/pipelines/go-elasticsearch-alerts

.. image:: https://img.shields.io/badge/godoc-reference-blue.svg
    :target: https://godoc.org/github.com/morningconsult/go-elasticsearch-alerts

.. image:: https://goreportcard.com/badge/github.com/morningconsult/go-elasticsearch-alerts
    :target: https://goreportcard.com/report/github.com/morningconsult/go-elasticsearch-alerts

Go Elasticsearch Alerts lets you create alerts on your Elasticsearch data. To
get started, just write a :ref:`configuration file <main-config-file>` and a
:ref:`rule <rule-configuration-file>` and you can receive alerts like the Slack
alert shown below when Go Elasticsearch Alerts finds new data matching the rule.

.. image:: ./_static/slack.png
   :class: shadowed-image

To try it out yourself, check out the :ref:`demonstration <demo>`.

The User Guide
--------------

This part of the documentation, begins with some background information
about the Go Elasticsearch Alerts project, then focuses on step-by-step
instructions for installing and using it.

.. toctree::
   :maxdepth: 2
   
   user/intro
   user/demo
   user/install
   user/setup
   user/usage