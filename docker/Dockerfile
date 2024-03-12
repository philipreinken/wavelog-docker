ARG PHP_VERSION="8.3"
FROM php:${PHP_VERSION}-apache

ARG WAVELOG_VERSION="1.0"

ADD https://github.com/mlocati/docker-php-extension-installer/releases/latest/download/install-php-extensions /usr/local/bin/

RUN <<EOF
    sed -ri 's/^Listen 80$/Listen 8080/g' /etc/apache2/ports.conf /etc/apache2/sites-available/*.conf
    sed -ri 's/VirtualHost \*:80/VirtualHost \*:8080/g' /etc/apache2/sites-enabled/000-default.conf
    a2enmod rewrite

    chmod +x /usr/local/bin/install-php-extensions
    install-php-extensions curl mbstring mysqli pdo_mysql xml zip
EOF

ADD --chown=www-data:www-data https://github.com/wavelog/wavelog.git#${WAVELOG_VERSION} /var/www/html

RUN cp /var/www/html/.htaccess.sample /var/www/html/.htaccess

EXPOSE 8080

USER www-data

WORKDIR /var/www/html