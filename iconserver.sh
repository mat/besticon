
# Simple wrapper to restart after crash,
# see http://stackoverflow.com/questions/696839
#
# Run with
#   nohup iconserver.sh > /var/log/iconserver.log &

export PORT=80
DAEMON="/usr/sbin/iconserver"

until $DAEMON; do
   echo "Server 'iconserver' crashed with exit code $?.  Respawning.." >&2
   sleep 1
done
