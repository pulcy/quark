# Quark: Cluster & instance creator

## Show all DNS records

```
quark dns records --domain pulcy.com
```

## Listing instances of a cluster

```
quark instances -p vultr --domain pulcy.com --name reg-c21
quark instances -p vultr reg-c21.pulcy.com
```

## Add a new instance to an existing cluster

```
quark create instance -p vultr --domain iggi.xyz --name alpha-tg2d
quark create instance -p vultr alpha-tg2d.iggi.xyz
```

## Removing an instance from an existing cluster

```
quark destroy instance -p vultr --domain iggi.xyz --name alpha-tg2d --prefix ldszw7sj
quark destroy instance -p vultr ldszw7sj.alpha-tg2d.iggi.xyz
```
