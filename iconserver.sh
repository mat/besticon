
# Simple wrapper to restart after crash,
# see http://stackoverflow.com/questions/696839
#
# Run with
#   nohup iconserver.sh > /var/log/iconserver.log &

DAEMON="/usr/sbin/iconserver --port 80"

until $DAEMON; do
   echo "Server 'iconserver' crashed with exit code $?.  Respawning.." >&2
   sleep 1
done
