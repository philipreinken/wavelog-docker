ARG PHP_VERSION="8.3"

FROM php:${PHP_VERSION}-apache as base

ADD --chmod=555 https://github.com/mlocati/docker-php-extension-installer/releases/latest/download/install-php-extensions /usr/local/bin/
RUN install-php-extensions curl mbstring mysqli pdo_mysql xml zip

FROM base as config

RUN <<EOF
    sed -ri 's/^Listen 80$/Listen 8080/g' /etc/apache2/ports.conf /etc/apache2/sites-available/*.conf
    sed -ri 's/VirtualHost \*:80/VirtualHost \*:8080/g' /etc/apache2/sites-enabled/000-default.conf
    a2enmod rewrite
EOF

FROM config

ARG WAVELOG_VERSION="1.0"

ADD --chown=www-data:www-data https://github.com/wavelog/wavelog.git#${WAVELOG_VERSION} /var/www/html

RUN cp /var/www/html/.htaccess.sample /var/www/html/.htaccess

EXPOSE 8080
USER www-data
WORKDIR /var/www/html