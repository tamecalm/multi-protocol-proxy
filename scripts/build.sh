#!/usr/bin/env bash


set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
source "${SCRIPT_DIR}/lib/colors.sh"
source "${SCRIPT_DIR}/lib/platform.sh"
source "${SCRIPT_DIR}/lib/utils.sh"

PROJECT_ROOT="$(get_project_root)"
cd "$PROJECT_ROOT"

BUILD_DIR="${PROJECT_ROOT}/build"

# Help Text
show_help() {
    echo ""
    echo "Usage: build.sh [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --all           Build for all supported platforms"
    echo "  --os <os>       Target OS (linux, darwin, windows)"
    echo "  --arch <arch>   Target architecture (amd64, arm64)"
    echo "  --clean         Clean build directory before building"
    echo "  --help          Show this help message"
    echo ""
    echo "Examples:"
    echo "  ./build.sh                    # Build for current platform"
    echo "  ./build.sh --all              # Build for all platforms"
    echo "  ./build.sh --os linux --arch md64"
    echo ""
}

build_for_platform() {
    local target_os="$1"
    local target_arch="$2"
    
    local binary_name=$(get_binary_name "$target_os" "$target_arch")
    local output_path="${BUILD_DIR}/${binary_name}"
    
    print_thinking "Building for ${target_os}-${target_arch}..."
    
    local start_time=$(get_timestamp)
    
    if GOOS="$target_os" GOARCH="$target_arch" go build -ldflags="-s -w" -o "$output_path" ./cmd/proxy 2>&1; then
        local end_time=$(get_timestamp)
        local elapsed=$(format_elapsed "$start_time" "$end_time")
        local size=$(get_file_size "$output_path")
        
        print_success "Built ${binary_name} (${size}) in ${elapsed}"
        return 0
    else
        print_error "Failed to build for ${target_os}-${target_arch}"
        return 1
    fi
}

build_all_platforms() {
    local platforms=$(get_supported_platforms)
    local success_count=0
    local fail_count=0
    
    print_section "Building All Platforms"
    
    while IFS= read -r platform; do
        local target_os="${platform%-*}"
        local target_arch="${platform#*-}"
        
        if build_for_platform "$target_os" "$target_arch"; then
            ((success_count++))
        else
            ((fail_count++))
        fi
    done <<< "$platforms"
    
    return $fail_count
}

main() {
    local build_all=false
    local target_os=""
    local target_arch=""
    local clean_first=false
    
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --all)
                build_all=true
                shift
                ;;
            --os)
                target_os="$2"
                shift 2
                ;;
            --arch)
                target_arch="$2"
                shift 2
                ;;
            --clean)
                clean_first=true
                shift
                ;;
            --help|-h)
                show_help
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
    
    print_banner "PROXY BUILD" "Cross-Platform Builder"
    
    if ! check_go_installed; then
        print_error "Go is not installed! Run ./scripts/install.sh first"
        exit 1
    fi
    
    if [[ -z "$target_os" ]]; then
        target_os=$(detect_os)
    fi
    
    if [[ -z "$target_arch" ]]; then
        target_arch=$(detect_arch)
    fi
    
    print_info "Go version: $(get_go_version)"
    print_info "Current platform: $(detect_os)-$(detect_arch)"
    
    if [[ "$clean_first" == true ]]; then
        print_thinking "Cleaning build directory..."
        clean_dir "$BUILD_DIR"
        print_success "Build directory cleaned"
    fi
    
    ensure_dir "$BUILD_DIR"
    
    local start_time=$(get_timestamp)
    local exit_code=0
    
    if [[ "$build_all" == true ]]; then
        if ! build_all_platforms; then
            exit_code=1
        fi
    else
        print_section "Building for ${target_os}-${target_arch}"
        
        if ! build_for_platform "$target_os" "$target_arch"; then
            exit_code=1
        fi
    fi
    
    local end_time=$(get_timestamp)
    local total_elapsed=$(format_elapsed "$start_time" "$end_time")
    
    print_section "Build Summary"
    
    if [[ $exit_code -eq 0 ]]; then
        print_group_start "Build Complete"
        print_group_item "Output Directory" "build/"
        print_group_item "Total Time" "$total_elapsed"
        print_group_end
        
        echo -e "  ${HI_BLACK}Built binaries:${RESET}"
        for file in "${BUILD_DIR}"/*; do
            if [[ -f "$file" ]]; then
                local name=$(basename "$file")
                local size=$(get_file_size "$file")
                echo -e "    ${GREEN}${ICON_SUCCESS}${RESET} ${CYAN}${name}${RESET} ${HI_BLACK}(${size})${RESET}"
            fi
        done
        echo ""
        
        echo -e "  ${HI_BLACK}▸${RESET} ${WHITE}Run ${CYAN}./scripts/run.sh${RESET}${WHITE} to start the proxy${RESET}"
        echo ""
    else
        print_error "Build failed! Check the errors above."
    fi
    
    exit $exit_code
}

main "$@"
