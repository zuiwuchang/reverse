Target="reverse"
Docker="king011/reverse"
Version="v0.0.1"
Dir=$(cd "$(dirname $BASH_SOURCE)/.." && pwd)
Platforms=(
    darwin/amd64
    windows/amd64
    linux/arm
    linux/amd64
)