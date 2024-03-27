# wavelog-docker

> unofficial container setup for [wavelog](https://github.com/wavelog/wavelog)

## Usage

There's an example `compose.yml` file [in the root of this repository](https://github.com/philipreinken/wavelog-docker),
you can use it as a starting point for your own setup. For a quick testrun, just run `docker compose up -d`, wait for
the command to finish and navigate to [http://localhost:8080](http://localhost:8080) in your browser.

## About

Published on hub.docker.com: https://hub.docker.com/r/philipreinken/wavelog

The image is currently based on the `php:8.2-apache` image and is rebuilt automatically on a regular basis for the
latest two minor versions of `wavelog`.

This is primarily intended for
experimentation and my personal setup, so use at your own risk. Contributions are welcome.
