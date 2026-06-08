#!/usr/bin/env bash

_supports_colors() {
    if [[ -n "${NO_COLOR:-}" ]]; then
        return 1
    fi
    if [[ -n "${FORCE_COLOR:-}" && "${FORCE_COLOR}" != "0" ]]; then
        return 0
    fi
    if [[ -t 1 ]]; then
        return 0
    fi
    return 1
}

if _supports_colors; then
    COLORS_ENABLED=true
else
    COLORS_ENABLED=false
fi

if [[ "${COLORS_ENABLED}" == true ]]; then
    RESET='\033[0m'
    
    BLACK='\033[0;30m'
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[0;33m'
    BLUE='\033[0;34m'
    MAGENTA='\033[0;35m'
    CYAN='\033[0;36m'
    WHITE='\033[0;37m'
    
    HI_BLACK='\033[0;90m'
    HI_RED='\033[0;91m'
    HI_GREEN='\033[0;92m'
    HI_YELLOW='\033[0;93m'
    HI_BLUE='\033[0;94m'
    HI_MAGENTA='\033[0;95m'
    HI_CYAN='\033[0;96m'
    HI_WHITE='\033[0;97m'
    
    BOLD='\033[1m'
    BOLD_RED='\033[1;31m'
    BOLD_GREEN='\033[1;32m'
    BOLD_YELLOW='\033[1;33m'
    BOLD_BLUE='\033[1;34m'
    BOLD_MAGENTA='\033[1;35m'
    BOLD_CYAN='\033[1;36m'
    BOLD_WHITE='\033[1;37m'
    
    BG_MAGENTA='\033[45m'
    BG_RED='\033[41m'
else
    RESET=''
    BLACK='' RED='' GREEN='' YELLOW='' BLUE='' MAGENTA='' CYAN='' WHITE=''
    HI_BLACK='' HI_RED='' HI_GREEN='' HI_YELLOW='' HI_BLUE='' HI_MAGENTA='' HI_CYAN='' HI_WHITE=''
    BOLD='' BOLD_RED='' BOLD_GREEN='' BOLD_YELLOW='' BOLD_BLUE='' BOLD_MAGENTA='' BOLD_CYAN='' BOLD_WHITE=''
    BG_MAGENTA='' BG_RED=''
fi

BOX_TOP_LEFT="╭"
BOX_TOP_RIGHT="╮"
BOX_BOTTOM_LEFT="╰"
BOX_BOTTOM_RIGHT="╯"
BOX_HORIZONTAL="─"
BOX_VERTICAL="│"
BOX_TEE_RIGHT="├"
BOX_TEE_LEFT="┤"


color_primary() {
    echo -e "${BOLD_MAGENTA}$1${RESET}"
}

color_secondary() {
    echo -e "${CYAN}$1${RESET}"
}

color_accent() {
    echo -e "${HI_RED}$1${RESET}"
}

color_accent_bright() {
    echo -e "${BOLD_RED}$1${RESET}"
}

color_success() {
    echo -e "${GREEN}$1${RESET}"
}

color_warn() {
    echo -e "${YELLOW}$1${RESET}"
}

color_error() {
    echo -e "${RED}$1${RESET}"
}

color_info() {
    echo -e "${HI_YELLOW}$1${RESET}"
}

color_muted() {
    echo -e "${HI_BLACK}$1${RESET}"
}

color_subtle() {
    echo -e "${WHITE}$1${RESET}"
}

color_heading() {
    echo -e "${BOLD_RED}$1${RESET}"
}

color_command() {
    echo -e "${BOLD_CYAN}$1${RESET}"
}

color_bold() {
    echo -e "${BOLD_WHITE}$1${RESET}"
}

ICON_SUCCESS="✓"
ICON_ERROR="✗"
ICON_WARNING="⚠"
ICON_INFO="ℹ"
ICON_SPINNER="◐"
ICON_DIAMOND="◆"
ICON_DIAMOND_EMPTY="◇"
ICON_ARROW="→"
ICON_BULLET="●"

_timestamp() {
    date "+%H:%M:%S"
}

print_banner() {
    local title="${1:-PROXY BUILD}"
    local subtitle="${2:-Build & Run Scripts}"
    local version="${3:-v1.0.0}"
    
    echo ""
    
    local badge="${BG_MAGENTA}${BOLD_WHITE} ◆ ${title} ${RESET}"
    local ver="${HI_BLACK}${version}${RESET}"
    
    local border_width=60
    local top_border="${HI_BLACK}${BOX_TOP_LEFT}$(printf '%*s' $border_width | tr ' ' "${BOX_HORIZONTAL}")${BOX_TOP_RIGHT}${RESET}"
    echo -e "$top_border"
    
    local title_content="${badge} ${ver}"
    local padding=$((border_width - 18 - ${#title} - ${#version}))
    echo -e "${HI_BLACK}${BOX_VERTICAL}${RESET}  ${title_content}$(printf '%*s' $padding)${HI_BLACK}${BOX_VERTICAL}${RESET}"
    
    local sub_padding=$((border_width - 2 - ${#subtitle}))
    echo -e "${HI_BLACK}${BOX_VERTICAL}${RESET}  ${WHITE}${subtitle}${RESET}$(printf '%*s' $sub_padding)${HI_BLACK}${BOX_VERTICAL}${RESET}"
    
    local bottom_border="${HI_BLACK}${BOX_BOTTOM_LEFT}$(printf '%*s' $border_width | tr ' ' "${BOX_HORIZONTAL}")${BOX_BOTTOM_RIGHT}${RESET}"
    echo -e "$bottom_border"
    echo ""
}

print_status() {
    local category="$1"
    local message="$2"
    local ts="${HI_BLACK}$(_timestamp)${RESET}"
    local icon=""
    local styled_msg=""
    
    case "$category" in
        success)
            icon="${GREEN}${ICON_SUCCESS}${RESET}"
            styled_msg="${GREEN}${message}${RESET}"
            ;;
        error)
            icon="${RED}${ICON_ERROR}${RESET}"
            styled_msg="${RED}${message}${RESET}"
            ;;
        warning|warn)
            icon="${YELLOW}${ICON_WARNING}${RESET}"
            styled_msg="${YELLOW}${message}${RESET}"
            ;;
        info)
            icon="${HI_YELLOW}${ICON_INFO}${RESET}"
            styled_msg="${WHITE}${message}${RESET}"
            ;;
        *)
            icon="${HI_BLACK}${ICON_BULLET}${RESET}"
            styled_msg="${WHITE}${message}${RESET}"
            ;;
    esac
    
    echo -e "${ts}  ${icon}  ${styled_msg}"
}

print_success() {
    print_status "success" "$1"
}

print_error() {
    print_status "error" "$1"
}

print_warn() {
    print_status "warning" "$1"
}

print_info() {
    print_status "info" "$1"
}

print_thinking() {
    local ts="${HI_BLACK}$(_timestamp)${RESET}"
    local spinner="${BOLD_MAGENTA}${ICON_SPINNER}${RESET}"
    echo -e "${ts}  ${spinner}  ${WHITE}$1${RESET}"
}

print_section() {
    local title="$1"
    local line_width=$((50 - ${#title}))
    echo ""
    echo -e "${HI_BLACK}──${RESET} ${BOLD_RED}${title}${RESET} ${HI_BLACK}$(printf '%*s' $line_width | tr ' ' '─')${RESET}"
}

print_separator() {
    echo -e "  ${HI_BLACK}$(printf '%*s' 56 | tr ' ' '─')${RESET}"
}

print_group_start() {
    local title="$1"
    local line_width=$((50 - ${#title}))
    echo ""
    echo -e "${HI_BLACK}${BOX_TOP_LEFT}──${RESET} ${BOLD_MAGENTA}${title}${RESET} ${HI_BLACK}$(printf '%*s' $line_width | tr ' ' "${BOX_HORIZONTAL}")${BOX_TOP_RIGHT}${RESET}"
}

print_group_item() {
    local label="$1"
    local value="$2"
    echo -e "${HI_BLACK}${BOX_VERTICAL}${RESET}  ${HI_BLACK}${label}:${RESET} ${CYAN}${value}${RESET}"
}

print_group_end() {
    echo -e "${HI_BLACK}${BOX_BOTTOM_LEFT}$(printf '%*s' 56 | tr ' ' "${BOX_HORIZONTAL}")${BOX_BOTTOM_RIGHT}${RESET}"
    echo ""
}

print_shutdown() {
    echo ""
    echo -e "  ${HI_BLACK}✋${RESET} ${HI_BLACK}Cancelled${RESET}"
}
