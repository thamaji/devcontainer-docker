#!/bin/bash
set -eu

args=()
mode=""
for arg in "${@}"; do
  case ${mode} in
    volume )
      path=$(realpath -m --relative-base="${CONTAINER_WORKSPACE}" "${arg%%:*}")
      if [[ "${path}" != /* ]]; then
        arg="${LOCAL_WORKSPACE}/${path}:${arg#*:}"
      else
        echo bind mount can only in workspace folder: ${path}
        exit 1
      fi
      mode=run
      ;;
    mount )
      declare -A params=()
      while read key value; do
        params[${key}]=${value}
      done < <(echo -n "${arg}" | awk 'BEGIN{RS=",";FS="=";OFS=" "}{print $1,$2}')

      if [ ${params[type]} = bind ]; then
        path=$(realpath -m --relative-base="${CONTAINER_WORKSPACE}" "${params[src]}")
        if [[ "${path}" != /* ]]; then
          params[src]="${LOCAL_WORKSPACE}/${path}"
        else
          echo bind mount can only in workspace folder: ${path}
          exit 1
        fi
      fi

      unset arg
      for key in "${!params[@]}"; do
        arg="${arg-}${arg+,}${key}=${params[${key}]}"
      done
      mode=run
      ;;
    run )
      case ${arg} in
        -v | --volume ) mode=volume ;;
        --mount ) mode=mount ;;
        -* ) ;;
        * ) mode="" ;;
      esac
      ;;
    * ) [ ${arg} == run ] && mode=run ;;
  esac
  args+=( "${arg}" )
done

exec /usr/bin/docker "${args[@]}"
