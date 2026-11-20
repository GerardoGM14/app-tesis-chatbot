# Preparación del VPS

Guía para dejar el VPS (Ubuntu 24.04 LTS) listo para recibir despliegues desde GitHub Actions.
Recomendado: 4 vCPU y 8 GB de RAM, porque SonarQube corre en el mismo servidor.

## 1. Sistema base

```bash
sudo apt update && sudo apt upgrade -y
sudo apt install -y ufw fail2ban unattended-upgrades

# Firewall: solo SSH y web
sudo ufw default deny incoming
sudo ufw allow OpenSSH
sudo ufw allow 80,443/tcp
sudo ufw allow 443/udp
sudo ufw enable
```

En `/etc/ssh/sshd_config`: `PasswordAuthentication no` y `PermitRootLogin no`. Luego `sudo systemctl restart ssh`.

## 2. Docker

```bash
curl -fsSL https://get.docker.com | sudo sh
```

> Docker publica puertos saltándose `ufw`. Por eso los compose de este repositorio solo publican los puertos de Caddy (80/443).

## 3. Usuario de despliegue y runner

```bash
sudo adduser --disabled-password --gecos "" deploy
sudo usermod -aG docker deploy
sudo mkdir -p /opt/plataforma && sudo chown deploy:deploy /opt/plataforma && sudo chmod 750 /opt/plataforma
```

> Pertenecer al grupo `docker` equivale en la práctica a ser root. Por eso el runner:
> - solo ejecuta jobs de `cd.yml`, que se disparan desde `main` y nunca desde pull requests;
> - corre con el usuario `deploy`, no con root;
> - conviene que el repositorio sea **privado**.

Instala el runner como `deploy` siguiendo **Settings › Actions › Runners › New self-hosted runner**. Al configurarlo, agrega la etiqueta `vps`:

```bash
./config.sh --url https://github.com/<usuario>/<repo> --token <token> --labels vps --name vps-plataforma
sudo ./svc.sh install deploy
sudo ./svc.sh start
```

## 4. Secretos de cada entorno

```bash
cd /opt/plataforma
cp /ruta/al/repo/deploy/env.example staging.env
cp /ruta/al/repo/deploy/env.example production.env
chmod 600 *.env
# Completa cada archivo con valores distintos:  openssl rand -hex 32
```

## 5. Caddy (TLS y enrutamiento)

Apunta los registros DNS A de `app.`, `staging.` y `sonar.` a la IP del VPS. Luego:

```bash
docker network create edge
cat > /opt/plataforma/caddy.env <<'EOF'
ACME_EMAIL=tu-correo@dominio
PROD_DOMAIN=app.tu-dominio
STAGING_DOMAIN=staging.tu-dominio
SONAR_DOMAIN=sonar.tu-dominio
EOF
cd /ruta/al/repo/deploy/caddy
docker compose --env-file /opt/plataforma/caddy.env up -d
```

## 6. SonarQube

```bash
echo "vm.max_map_count=524288" | sudo tee /etc/sysctl.d/99-sonarqube.conf
echo "fs.file-max=131072"      | sudo tee -a /etc/sysctl.d/99-sonarqube.conf
sudo sysctl --system

echo "SONAR_DB_PASSWORD=$(openssl rand -hex 24)" > /opt/plataforma/sonarqube.env
chmod 600 /opt/plataforma/sonarqube.env
cd /ruta/al/repo/deploy/sonarqube
docker compose --env-file /opt/plataforma/sonarqube.env up -d
```

En `https://sonar.tu-dominio`:

1. Entra con `admin` / `admin` y cambia la contraseña.
2. Crea el proyecto con la clave `plataforma-atencion`.
3. Genera un token de análisis y guárdalo en GitHub como secret `SONAR_TOKEN`.
4. En **Quality Gates**, crea uno con las condiciones de [docs/04-calidad](../docs/04-calidad/estrategia-de-pruebas.md#quality-gate-de-sonarqube) y asígnalo al proyecto.

## 7. Backups

```bash
sudo mkdir -p /var/backups/plataforma && sudo chown deploy:deploy /var/backups/plataforma
crontab -u deploy -e
# 15 3 * * * /home/deploy/actions-runner/_work/<repo>/<repo>/deploy/scripts/backup.sh production
```

## Cómo funciona un despliegue

[`scripts/deploy.sh`](scripts/deploy.sh) recibe el entorno y el tag (SHA del commit):

1. Descarga las imágenes de GHCR.
2. Ejecuta `migrate`, que aplica migraciones, activa el rol `plataforma_app` y crea el superadmin si falta.
3. Actualiza los servicios y espera sus healthchecks.
4. Si algo falla, vuelve a la última versión buena (`/opt/plataforma/<entorno>.last-good`).

Las migraciones solo avanzan: cada una debe ser compatible con la versión anterior del backend para que el rollback sea seguro.
