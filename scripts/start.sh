#!/usr/bin/env bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/colors.sh"
source "${SCRIPT_DIR}/lib/platform.sh"
source "${SCRIPT_DIR}/lib/utils.sh"

PROJECT_ROOT="$(get_project_root)"
cd "$PROJECT_ROOT"

show_help() {
    echo ""
    echo "Usage: start.sh [COMMAND] [OPTIONS]"
    echo ""
    echo "Commands:"
    echo "  install         Install dependencies"
    echo "  build           Build for current platform"
    echo "  build-all       Build for all platforms"
    echo "  run             Run the proxy"
    echo "  dev             Build and run (development)"
    echo "  info            Show platform information"
    echo ""
    echo "Options:"
    echo "  --help, -h      Show this help message"
    echo ""
    echo "Examples:"
    echo "  ./start.sh                  # Show interactive menu"
    echo "  ./start.sh install          # Install dependencies"
    echo "  ./start.sh build            # Build for current platform"
    echo "  ./start.sh dev              # Build and run"
    echo ""
}

cmd_install() {
    exec "${SCRIPT_DIR}/install.sh" "$@"
}

cmd_build() {
    exec "${SCRIPT_DIR}/build.sh" "$@"
}

cmd_build_all() {
    exec "${SCRIPT_DIR}/build.sh" --all "$@"
}

cmd_run() {
    exec "${SCRIPT_DIR}/run.sh" "$@"
}

cmd_dev() {
    "${SCRIPT_DIR}/build.sh"
    echo ""
    exec "${SCRIPT_DIR}/run.sh" --dev
}

cmd_info() {
    print_banner "SYSTEM INFO" "Platform Information"
    print_platform_summary
    # Check Go
    print_group_start "Environment"
    if check_go_installed; then
        print_group_item "Go" "$(get_go_version)"
    else
        print_group_item "Go" "Not installed"
    fi
    
    local build_count=$(find "${PROJECT_ROOT}/build" -type f 2>/dev/null | wc -l)
    print_group_item "Built Binaries" "${build_count} found"
    print_group_end
    
    if [[ $build_count -gt 0 ]]; then
        echo -e "  ${HI_BLACK}Available binaries:${RESET}"
        for file in "${PROJECT_ROOT}/build"/*; do
            if [[ -f "$file" ]]; then
                local name=$(basename "$file")
                local size=$(get_file_size "$file")
                echo -e "    ${CYAN}${name}${RESET} ${HI_BLACK}(${size})${RESET}"
            fi
        done
        echo ""
    fi
}

show_menu() {
    print_banner "MULTI-PROTOCOL PROXY" "Build & Run Scripts"
    
    local os=$(detect_os)
    local arch=$(detect_arch)
    print_info "Platform: ${os}-${arch}"
    print_info "Go version: $(get_go_version 2>/dev/null || echo 'not installed')"
    
    print_section "Select an Option"
    echo ""
    
    echo -e "  ${CYAN}1)${RESET} ${WHITE}Install Dependencies${RESET}   ${HI_BLACK}— Download Go modules${RESET}"
    echo -e "  ${CYAN}2)${RESET} ${WHITE}Build${RESET}                  ${HI_BLACK}— Build for current platform${RESET}"
    echo -e "  ${CYAN}3)${RESET} ${WHITE}Build All${RESET}              ${HI_BLACK}— Build for all platforms${RESET}"
    echo -e "  ${CYAN}4)${RESET} ${WHITE}Run${RESET}                    ${HI_BLACK}— Run the proxy${RESET}"
    echo -e "  ${CYAN}5)${RESET} ${WHITE}Build & Run${RESET}            ${HI_BLACK}— Build then run (dev mode)${RESET}"
    echo -e "  ${CYAN}6)${RESET} ${WHITE}Platform Info${RESET}          ${HI_BLACK}— Show system information${RESET}"
    echo -e "  ${CYAN}q)${RESET} ${WHITE}Quit${RESET}"
    echo ""
    
    read -r -p "  Select option [1-6, q]: " choice
    
    case "$choice" in
        1)
            echo ""
            cmd_install
            ;;
        2)
            echo ""
            cmd_build
            ;;
        3)
            echo ""
            cmd_build_all
            ;;
        4)
            echo ""
            cmd_run
            ;;
        5)
            echo ""
            cmd_dev
            ;;
        6)
            echo ""
            cmd_info
            ;;
        q|Q)
            echo ""
            print_info "Goodbye!"
            exit 0
            ;;
        *)
            print_error "Invalid option: $choice"
            exit 1
            ;;
    esac
}

main() {
    trap 'print_shutdown; exit 0' INT
    
    if [[ $# -eq 0 ]]; then
        show_menu
        exit 0
    fi
    
    # Parse command
    local command="$1"
    shift
    
    case "$command" in
        install)
            cmd_install "$@"
            ;;
        build)
            cmd_build "$@"
            ;;
        build-all)
            cmd_build_all "$@"
            ;;
        run)
            cmd_run "$@"
            ;;
        dev)
            cmd_dev "$@"
            ;;
        info)
            cmd_info "$@"
            ;;
        --help|-h|help)
            show_help
            ;;
        *)
            print_error "Unknown command: $command"
            show_help
            exit 1
            ;;
    esac
}

main "$@"
