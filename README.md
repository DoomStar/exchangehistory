# Programming test Simon Spiridonov 

### run the docker-compose

`mkdir data/grafana ; chmod 777 data/grafana; docker-compose up -d --build`

After launch it will: 

* Download images and build the application
* Run `influxdb` and `grafana`
* Application will run automatically after `influxdb` is started and healthy
* Application will gather history and store it to `influxdb` using `./main history --from 2018-01-01 --to 2019-01-01` as run command
* `CronD` in `app` container will run `./main update` every 6 hours and application will try to push new data from `https://api.exchangeratesapi.io/latest` ot `InfluxDB` 
* You can open http://localhost:3000 in your browser to access `grafana`
* Use `admin`/`admin` credentions
* Navigate to "Manage" and then to "Currency" dashboard
* You will see USD currency exchange rate
* You can change currency in upper left corner
* To add more data to influxdb run commands:
  * docker exec -i app ./main update
  * refresh grafana page to see the difference
  * docker exec -i app ./main history --from 2019-01-01 --to 2019-04-01
  * refresh grafana page to see the difference

# Tools

### InfluxDB

It can efficiently store values relative to timestamps

### Golang

I'm writing on this language for ~2 months and I enjoy learning it. So I decided to do this task to learn more. I prefer this language to python for workng with data.

### Grafana

It's so easy to display data from InfluxDB using grafana :)

# Kubernetes

```
docker build -f Dockerfile_grafana -t docker-registry.ppdev.ru/app/grafana:latest .
docker push docker-registry.ppdev.ru/app/grafana:latest
docker tag exchangehistory_app docker-registry.ppdev.ru/app/app:1
docker push docker-registry.ppdev.ru/app/app:1
helm upgrade ./helm --install --kube-context stage --namespace app --set-file config_ini=config/config.ini --set build_id=1
```

After this application wil be deployed and filled with same history as on local deployment with docker-compose. 
You can add to `/etc/hosts`:

```
<your ingress ip> grafana.app.check24dev.de
```

And you will be able to use grafana.

To test cronjob run this:

```
kubectl --context stage --namespace app create job app-upd --from=cronjobs/app
``` 

