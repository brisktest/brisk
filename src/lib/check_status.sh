echo $1
if [ -z `docker-compose ps -q $1` ] || [ -z `docker ps -q --no-trunc | grep $(docker-compose ps -q $1)` ]; then
  echo 'Not running.''
  exit 1
else
  echo 'Yes, running.'
  exit 0 
fi