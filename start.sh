# Fix Nginx config
sed "s/<HOSTNAME>/$HOSTNAME/g" /etc/nginx/conf.d/default.conf > /var/tmp/default.conf
cat /var/tmp/default.conf > /etc/nginx/conf.d/default.conf

nginx
/api/api &
node /app/server.js &
wait -n
