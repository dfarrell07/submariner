#!/bin/bash
# Quick check of repository access for Submariner RPM dependencies
#
# Tests if repos are accessible (subscription entitlements).
# For full package verification, use: verify-packages.sh
#
# Requires: curl, entitlement certs in /etc/pki/entitlement/

set -euo pipefail

# Colors
R='\033[31m' G='\033[32m' Y='\033[33m' B='\033[1m' N='\033[0m'

# Find entitlement cert
CERT=$(find /etc/pki/entitlement -name "*.pem" ! -name "*-key.pem" 2>/dev/null | head -1)
KEY=${CERT%.pem}-key.pem
[[ -f "$CERT" ]] || { echo "No entitlement certs. Run: sudo subscription-manager register"; exit 1; }

# Check repo access with CA cert (returns 0=accessible, 1=blocked)
check() {
    local code
    code=$(curl -s -o /dev/null -w "%{http_code}" --cert "$CERT" --key "$KEY" \
        --cacert /etc/rhsm/ca/redhat-uep.pem "$1" 2>/dev/null) || code=000
    [[ $code == 200 ]]
}

check_public() {
    local code
    code=$(curl -s -o /dev/null -w "%{http_code}" "$1" 2>/dev/null) || code=000
    [[ $code == 200 ]]
}

# Repo URLs
RHEL_GA="https://cdn.redhat.com/content/dist/rhel9/9"
RHEL_BETA="https://cdn.redhat.com/content/beta/rhel9/9"
RHEL_EUS="https://cdn.redhat.com/content/eus/rhel9/9.4"
FDP="https://cdn.redhat.com/content/dist/layered/rhel9"
UBI="https://cdn-ubi.redhat.com/content/public/ubi/dist/ubi9/9"

echo -e "${B}Submariner RPM Dependency Status${N}"
echo "================================="
echo
echo -e "${B}Component  Arch     Repository        Status${N}"
echo "---------- -------- ----------------- ------"

# gateway: libreswan from RHEL 9
# x86_64/aarch64: GA (dist), ppc64le: beta, s390x: EUS 9.4
for arch in x86_64 aarch64; do
    printf "gateway    %-8s RHEL 9 GA         " "$arch"
    if check "$RHEL_GA/$arch/baseos/os/repodata/repomd.xml"; then
        echo -e "${G}OK${N}"
    else
        echo -e "${R}403${N}"
    fi
done

printf "gateway    %-8s RHEL 9 beta       " "ppc64le"
if check "$RHEL_BETA/ppc64le/baseos/os/repodata/repomd.xml"; then
    echo -e "${G}OK${N}"
else
    echo -e "${R}403${N}"
fi

printf "gateway    %-8s RHEL 9 EUS 9.4    " "s390x"
if check "$RHEL_EUS/s390x/baseos/os/repodata/repomd.xml"; then
    echo -e "${G}OK${N}"
else
    echo -e "${R}403${N}"
fi

echo

# route-agent: openvswitch from fast-datapath (only x86_64/aarch64 available)
for arch in x86_64 aarch64; do
    printf "route-agent %-7s fast-datapath     " "$arch"
    if check "$FDP/$arch/fast-datapath/os/repodata/repomd.xml"; then
        echo -e "${G}OK${N}"
    else
        echo -e "${R}403${N}"
    fi
done

printf "route-agent %-7s fast-datapath     " "ppc64le"
echo -e "${Y}N/A${N} (not available)"

printf "route-agent %-7s fast-datapath     " "s390x"
echo -e "${Y}N/A${N} (not available)"

echo

# globalnet: iptables-nft from UBI (public, all arches)
for arch in x86_64 aarch64 ppc64le s390x; do
    printf "globalnet  %-8s UBI (public)      " "$arch"
    if check_public "$UBI/$arch/baseos/os/repodata/repomd.xml"; then
        echo -e "${G}OK${N}"
    else
        echo -e "${R}N/A${N}"
    fi
done

echo
echo -e "${B}Legend:${N} ${G}OK${N}=accessible  ${R}403${N}=blocked  ${Y}N/A${N}=arch not in repo"
