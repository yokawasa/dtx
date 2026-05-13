package cli

func completionScript(shell string) string {
	switch shell {
	case "bash":
		return bashCompletion
	case "zsh":
		return zshCompletion
	case "fish":
		return fishCompletion
	default:
		return ""
	}
}

const bashCompletion = `# bash completion for dtx
_dtx_envs() {
  local dtx_home
  dtx_home="${DTX_HOME:-$HOME/.dtx}"
  if [ ! -d "$dtx_home/envs" ]; then
    return
  fi
  command ls -1 "$dtx_home/envs" 2>/dev/null | command sed -n 's/\.enc$//p'
}

_dtx() {
  local cur cmd arg
  local used_verbose=0
  local env_set=0
  cur="${COMP_WORDS[COMP_CWORD]}"

  if [ "$COMP_CWORD" -eq 1 ]; then
    COMPREPLY=( $(compgen -W "use current ls run edit completion" -- "$cur") )
    return
  fi

  cmd="${COMP_WORDS[1]}"
  case "$cmd" in
    use|edit)
      COMPREPLY=( $(compgen -W "$(_dtx_envs)" -- "$cur") )
      return
      ;;
    current|ls)
      return
      ;;
    completion)
      COMPREPLY=( $(compgen -W "bash zsh fish" -- "$cur") )
      return
      ;;
    run)
      for ((i = 2; i < COMP_CWORD; i++)); do
        arg="${COMP_WORDS[i]}"
        if [ "$arg" = "--" ]; then
          return
        fi
        if [ "$arg" = "--verbose" ]; then
          used_verbose=1
          continue
        fi
        if [ -n "$arg" ]; then
          env_set=1
        fi
      done

      local choices=""
      if [ "$used_verbose" -eq 0 ]; then
        choices="$choices --verbose"
      fi
      choices="$choices --"
      if [ "$env_set" -eq 0 ]; then
        choices="$choices $(_dtx_envs)"
      fi

      COMPREPLY=( $(compgen -W "$choices" -- "$cur") )
      return
      ;;
  esac
}

complete -o bashdefault -o default -F _dtx dtx
`

const zshCompletion = `#compdef dtx

if ! whence compdef >/dev/null 2>&1; then
  autoload -Uz compinit
  compinit
fi

_dtx_env_names() {
  local dtx_home
  dtx_home=${DTX_HOME:-$HOME/.dtx}
  print -l -- ${${(f)"$(command ls -1 "$dtx_home/envs" 2>/dev/null)"}%.enc}
}

_dtx_envs() {
  local -a envs
  envs=(${(f)"$(_dtx_env_names)"})
  compadd -- "$@" $envs
}

_dtx() {
  local -a commands
  commands=(
    'use:select the current env'
    'current:print the current env'
    'ls:list available envs'
    'run:run a command with an env'
    'edit:create or edit an env'
    'completion:generate shell completion'
  )

  if (( CURRENT == 2 )); then
    _describe 'command' commands
    return
  fi

  case "${words[2]}" in
    use|edit)
      _dtx_envs
      return
      ;;
    current|ls)
      return
      ;;
    completion)
      compadd -- bash zsh fish
      return
      ;;
    run)
      local i arg
      local used_verbose=0
      local env_set=0
      for (( i = 3; i < CURRENT; i++ )); do
        arg="${words[i]}"
        if [[ "$arg" == "--" ]]; then
          return
        fi
        if [[ "$arg" == "--verbose" ]]; then
          used_verbose=1
          continue
        fi
        if [[ -n "$arg" ]]; then
          env_set=1
        fi
      done

      local -a choices
      choices=(--)
      if (( ! used_verbose )); then
        choices+=(--verbose)
      fi
      if (( ! env_set )); then
        local -a envs
        envs=(${(f)"$(_dtx_env_names)"})
        choices+=($envs)
      fi
      compadd -- $choices
      return
      ;;
  esac
}

compdef _dtx dtx
`

const fishCompletion = `function __dtx_envs
    set -l dtx_home
    if set -q DTX_HOME
        set dtx_home $DTX_HOME
    else
        set dtx_home $HOME/.dtx
    end
    if test -d "$dtx_home/envs"
        for file in "$dtx_home"/envs/*.enc
            if test -f "$file"
                string replace -r '\.enc$' '' -- (path basename "$file")
            end
        end
    end
end

function __dtx_run_needs_separator
    not contains -- -- (commandline -opc)
end

complete -c dtx -f
complete -c dtx -n '__fish_use_subcommand' -a 'use current ls run edit completion'
complete -c dtx -n '__fish_seen_subcommand_from use' -a '(__dtx_envs)'
complete -c dtx -n '__fish_seen_subcommand_from edit' -a '(__dtx_envs)'
complete -c dtx -n '__fish_seen_subcommand_from completion' -a 'bash zsh fish'
complete -c dtx -n '__fish_seen_subcommand_from run; and not __fish_seen_argument -l verbose' -l verbose
complete -c dtx -n '__fish_seen_subcommand_from run; and __dtx_run_needs_separator' -a '(__dtx_envs)'
complete -c dtx -n '__fish_seen_subcommand_from run; and __dtx_run_needs_separator' -a --
`
