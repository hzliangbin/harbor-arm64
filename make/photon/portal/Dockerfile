FROM photon:3.0

COPY build_dir/dist /usr/share/nginx/html
COPY build_dir/swagger.yaml /usr/share/nginx/html
COPY build_dir/swagger.json /usr/share/nginx/html
COPY build_dir/LICENSE /usr/share/nginx/html

COPY make/photon/portal/nginx.conf /etc/nginx/nginx.conf

RUN tdnf install -y nginx sudo >> /dev/null \
    && ln -sf /dev/stdout /var/log/nginx/access.log \
    && ln -sf /dev/stderr /var/log/nginx/error.log \
    && groupadd -r -g 10000 nginx && useradd --no-log-init -r -g 10000 -u 10000 nginx \
    && chown -R nginx:nginx /etc/nginx \
    && tdnf clean all

EXPOSE 8080
VOLUME /var/cache/nginx /var/log/nginx /run

STOPSIGNAL SIGQUIT

HEALTHCHECK CMD curl --fail -s http://127.0.0.1:8080 || exit 1
USER nginx
CMD ["nginx", "-g", "daemon off;"]

