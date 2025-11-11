# FROM --platform=linux/amd64 php:8.2-cli

# # Install system dependencies and PHP extensions
# RUN apt-get update && apt-get install -y \
#     git \
#     unzip \
#     zlib1g-dev \
#     && rm -rf /var/lib/apt/lists/*

# # Install gRPC and protobuf extensions
# RUN pecl install grpc-1.57.0 protobuf \
#     && docker-php-ext-enable grpc protobuf

# # Install Composer
# COPY --from=composer:latest /usr/bin/composer /usr/bin/composer

# WORKDIR /app

# # Copy composer files
# COPY examples/plugins/php/composer.json examples/plugins/php/composer.lock* ./

# # Copy generated protobuf files (needed for autoload)
# COPY proto/generated/php/ /proto/generated/php/

# # Install PHP dependencies  
# RUN composer install --no-interaction --prefer-dist

# # Copy application files
# COPY examples/plugins/php/src/ /app/src/

# CMD ["php", "src/HelloPlugin.php"]

FROM --platform=linux/amd64 debian:bookworm-slim

RUN apt-get update && apt-get install -y \
    libxml2 \
    libssl3 \
    git \
    unzip \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Install Composer
COPY --from=composer:latest /usr/bin/composer /usr/bin/composer

# Copy generated protobuf files
COPY proto/generated/php/ /proto/generated/php/

# Copy composer files
COPY examples/plugins/php/composer.json examples/plugins/php/composer.lock* ./

CMD ["/usr/local/php7/bin/php", "src/HelloPlugin.php"]