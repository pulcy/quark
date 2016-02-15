# Quark: Cluster & instance creator

## Show all DNS records

```
quark dns records --domain pulcy.com
```

## Listing instances of a cluster

```
quark instance list -p vultr --domain pulcy.com --name reg-c21
quark instance list -p vultr reg-c21.pulcy.com
```

## Add a new instance to an existing cluster

```
quark instance create -p vultr --domain iggi.xyz --name alpha-tg2d
quark instance create -p vultr alpha-tg2d.iggi.xyz
```

## Removing an instance from an existing cluster

```
quark instance destroy -p vultr --domain iggi.xyz --name alpha-tg2d --prefix ldszw7sj
quark instance destroy -p vultr ldszw7sj.alpha-tg2d.iggi.xyz
```
