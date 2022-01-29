#!/usr/bin/env bash
#=================================================
#  System Required: CentOS/Debian/ArchLinux with Systemd Support
#  Description: ServerStatus-goclient
#  Version: v1.0.1
#  Author: Kagurazaka Mizuki && RvvcIm
#=================================================

Green_font_prefix="\033[32m" && Red_font_prefix="\033[31m" && Red_background_prefix="\033[41;37m" && Font_color_suffix="\033[0m"
Info="${Green_font_prefix}[信息]${Font_color_suffix}"
Error="${Red_font_prefix}[错误]${Font_color_suffix}"
Tip="${Green_font_prefix}[注意]${Font_color_suffix}"

function check_sys() {
  if [[ -f /etc/redhat-release ]]; then
    release="centos"
  elif grep -q -E -i "debian|ubuntu" /etc/issue; then
    release="debian"
  elif grep -q -E -i "centos|red hat|redhat" /etc/issue; then
    release="centos"
  elif grep -q -E -i "Arch|Manjaro" /etc/issue; then
    release="arch"
  elif grep -q -E -i "debian|ubuntu" /proc/version; then
    release="debian"
  elif grep -q -E -i "centos|red hat|redhat" /proc/version; then
    release="centos"
  else
    echo -e "Status Client 暂不支持该 Linux 发行版"
    exit 1
  fi
  bit=$(uname -m)
}

function check_pid() {
  PID=$(pgrep -f "status-client")
}

function install_dependencies() {
  case ${release} in
  centos)
    yum install -y wget curl
    ;;
  debian)
    apt-get update -y
    apt-get install -y wget curl
    ;;
  arch)
    pacman -Syu --noconfirm wget curl
    ;;
  *)
    exit 1
    ;;
  esac
}

function input_dsn() {
  echo -e "${Info} 请输入服务端的 DSN, 格式为 “username:password@masterip(:port)”"
  read -re dsn
}

service_conf=/usr/lib/systemd/system/status-client.service
#service_conf=test.conf

function write_service() {
  echo -e "${Info} 写入systemd配置中"
  cat >${service_conf} <<-EOF
[Unit]
Description=ServerStatus-Client
Documentation=https://github.com/cokemine/ServerStatus-goclient
After=network.target

[Service]
ExecStart=/usr/local/ServerStatus/client/status-client -dsn="${dsn}"
ExecReload=/bin/kill -HUP $MAINPID
Restart=on-failure

[Install]
WantedBy=multi-user.target
EOF

}

function enable_service() {
  write_service
  systemctl enable status-client
  systemctl start status-client
  check_pid
  if [[ -n ${PID} ]]; then
    echo -e "${Info} Status Client 启动成功！"
  else
    echo -e "${Error} Status Client 启动失败！"
  fi
}

function restart_service() {
  write_service
  systemctl daemon-reload
  systemctl restart status-client
  check_pid
  if [[ -n ${PID} ]]; then
    echo -e "${Info} Status Client 启动成功！"
  else
    echo -e "${Error} Status Client 启动失败！"
  fi
}

function reset_config() {
  restart_service
}

function install_client() {
  case ${bit} in
  x86_64)
    arch=amd64
    ;;
  i386)
    arch=386
    ;;
  aarch64 | aarch64_be | arm64 | armv8b | armv8l)
    arch=arm64
    ;;
  arm | armv6l | armv7l | armv5tel | armv5tejl)
    arch=arm
    ;;
  mips | mips64)
    arch=mips
    ;;
  *)
    exit 1
    ;;
  esac
  echo -e "${Info} 下载 ${arch} 二进制文件"
  mkdir -p /usr/local/ServerStatus/client/
  cd /tmp && wget "https://github.com/cokemine/ServerStatus-goclient/releases/latest/download/status-client_linux_${arch}.tar.gz"
  tar -zxvf "status-client_linux_${arch}.tar.gz" status-client
  mv status-client /usr/local/ServerStatus/client/
  chmod +x /usr/local/ServerStatus/client/status-client
  enable_service
}

function auto_install() {
  dsn=$(echo ${*})
  install_client
}

function uninstall_client() {
  systemctl stop status-client
  systemctl disable status-client
  rm -rf /usr/local/ServerStatus/client/
  rm -rf /usr/lib/systemd/system/status-client.service
}

check_sys
case "$1" in
uninstall|uni)
  uninstall_client
  ;;
reset_conf|re)
  input_dsn
  reset_config
  ;;
-dsn)
  shift 1
  install_dependencies
  auto_install ${*}
  ;;
*)
  install_dependencies
  input_dsn
  install_client
  ;;
esac
