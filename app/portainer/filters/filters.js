import moment from 'moment';
import _ from 'lodash-es';
import filesize from 'filesize';
import { Cloud } from 'lucide-react';

import Kube from '@/assets/ico/kube.svg?c';
import DockerIcon from '@/assets/ico/vendor/docker-icon.svg?c';
import MicrosoftIcon from '@/assets/ico/vendor/microsoft-icon.svg?c';
import NomadIcon from '@/assets/ico/vendor/nomad-icon.svg?c';
import { EnvironmentType } from '@/react/portainer/environments/types';

export function truncateLeftRight(text, max, left, right) {
  max = isNaN(max) ? 50 : max;
  left = isNaN(left) ? 25 : left;
  right = isNaN(right) ? 25 : right;

  if (text.length <= max) {
    return text;
  } else {
    return text.substring(0, left) + '[...]' + text.substring(text.length - right, text.length);
  }
}

export function stripProtocol(url) {
  return url.replace(/.*?:\/\//g, '');
}

export function humanize(bytes, round, base) {
  if (!round) {
    round = 1;
  }
  if (!base) {
    base = 10;
  }
  if (bytes || bytes === 0) {
    return filesize(bytes, { base: base, round: round });
  }
}

export const TIME_FORMAT = 'YYYY-MM-DD HH:mm:ss';

export function isoDateFromTimestamp(timestamp) {
  return moment.unix(timestamp).format(TIME_FORMAT);
}

export function isoDate(date) {
  return moment(date).format(TIME_FORMAT);
}

export function parseIsoDate(date) {
  return moment(date, TIME_FORMAT).toDate();
}

export function formatDate(date, strFormat = 'YYYY-MM-DD HH:mm:ss Z') {
  return moment(date, strFormat).format(TIME_FORMAT);
}

export function getPairKey(pair, separator) {
  if (!pair.includes(separator)) {
    return pair;
  }

  return pair.slice(0, pair.indexOf(separator));
}

export function getPairValue(pair, separator) {
  if (!pair.includes(separator)) {
    return '';
  }

  return pair.slice(pair.indexOf(separator) + 1);
}

export function ipAddress(ip) {
  return ip.slice(0, ip.indexOf('/'));
}

export function arrayToStr(arr, separator) {
  if (arr) {
    return _.join(arr, separator);
  }
  return '';
}

export function labelsToStr(arr, separator) {
  if (arr) {
    return _.join(
      arr.map((item) => item.key + ':' + item.value),
      separator
    );
  }
  return '';
}

export function endpointTypeName(type) {
  switch (type) {
    case EnvironmentType.Docker:
      return 'Docker';
    case EnvironmentType.AgentOnDocker:
    case EnvironmentType.AgentOnKubernetes:
      return 'Agent';
    case EnvironmentType.Azure:
      return 'Azure ACI';
    case EnvironmentType.KubernetesLocal:
      return 'Kubernetes';
    case EnvironmentType.EdgeAgentOnDocker:
    case EnvironmentType.EdgeAgentOnKubernetes:
    case EnvironmentType.EdgeAgentOnNomad:
      return 'Edge Agent';
    default:
      throw new Error(`type ${type}-${EnvironmentType[type]} is not supported`);
  }
}

export function environmentTypeIcon(type) {
  switch (type) {
    case EnvironmentType.Azure:
      return MicrosoftIcon;
    case EnvironmentType.EdgeAgentOnDocker:
      return Cloud;
    case EnvironmentType.AgentOnKubernetes:
    case EnvironmentType.EdgeAgentOnKubernetes:
    case EnvironmentType.KubernetesLocal:
      return Kube;
    case EnvironmentType.AgentOnDocker:
    case EnvironmentType.Docker:
      return DockerIcon;
    case EnvironmentType.EdgeAgentOnNomad:
      return NomadIcon;
    default:
      throw new Error(`type ${type}-${EnvironmentType[type]} is not supported`);
  }
}

export function licenseTypeName(type) {
  switch (type) {
    case 1:
      return 'Trial';
    case 2:
      return 'Subscription';
    case 3:
      return 'Free';
    case 4:
      return 'Personal';
    case 5:
      return 'Starter';
    default:
      throw new Error(`License type ${type} is not supported`);
  }
}

export function truncate(text, length, end) {
  if (isNaN(length)) {
    length = 10;
  }

  if (end === undefined) {
    end = '...';
  }

  if (text.length <= length || text.length - end.length <= length) {
    return text;
  } else {
    return String(text).substring(0, length - end.length) + end;
  }
}
