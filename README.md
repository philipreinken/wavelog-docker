# wavelog-docker

> unofficial container setup for [wavelog](https://github.com/wavelog/wavelog)

## Usage

There's an example `compose.yml` file [in the root of this repository](https://github.com/philipreinken/wavelog-docker),
you can use it as a starting point for your own setup. For a quick testrun, just run `docker compose up -d`, wait for
the command to finish and navigate to [http://localhost:8080](http://localhost:8080) in your browser.

## About

The image is currently based on the `php:8.2-apache` image and is rebuilt automatically on a regular basis for the
latest two minor versions of `wavelog`.

There's no connection to the [wavelog](https://github.com/wavelog/wavelog) project, this is primarily intended for
experimentation and my own setup, so use at your own risk. Contributions are welcome.
