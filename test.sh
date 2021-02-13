if [ -f "/etc/secrets/testdb/password" ]; then
    echo "File /etc/secrets/testdb/password exists"
else
    echo "File /etc/secrets/testdb/password does not exist"
fi
