import { v4 as uuidv4 } from 'uuid';
import { KubernetesPortainerApplicationNote } from '../models/application/models';
import { KubernetesSystem_AnnotationsToSkip } from '../models/history/models';

export function KubernetesFormValueAnnotation() {
  return {
    ID: '',
    Key: '',
    Value: '',
  };
}

class KubernetesAnnotationsUtils {
  static apiToFormValueAnnotations(annotations) {
    const res = [];
    if (annotations) {
      Object.keys(annotations).forEach((k) => {
        if (!KubernetesSystem_AnnotationsToSkip[k] && k !== KubernetesPortainerApplicationNote) {
          const annotation = new KubernetesFormValueAnnotation();
          annotation.Key = k;
          annotation.Value = annotations[k];
          annotation.ID = uuidv4();

          res.push(annotation);
        }
      });
    }
    return res;
  }

  static formValuesToKubeAnnotations(formValues) {
    const annotations = {};
    if (!formValues.Annotations) {
      return undefined;
    }
    formValues.Annotations.forEach((a) => {
      annotations[a.Key] = a.Value;
    });
    return annotations;
  }

  static validateAnnotations(annotations) {
    const duplicatedAnnotations = [];
    const annotationsErrors = {};
    const re = /^([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9]$/;
    annotations.forEach((a, i) => {
      if (!a.Key) {
        annotationsErrors[`annotations.key[${i}]`] = 'Key is required.';
      } else if (duplicatedAnnotations.includes(a.Key)) {
        annotationsErrors[`annotations.key[${i}]`] = 'Key is a duplicate of an existing one.';
      } else {
        const key = a.Key.split('/');
        if (key.length > 2) {
          annotationsErrors[`annotations.key[${i}]`] = 'Two segments are allowed, separated by a slash (/): a prefix (optional) and a name.';
        } else if (key.length === 2) {
          if (key[0].length > 253) {
            annotationsErrors[`annotations.key[${i}]`] = "Prefix (before the slash) can't exceed 253 characters.";
          } else if (key[1].length > 63) {
            annotationsErrors[`annotations.key[${i}]`] = "Name (after the slash) can't exceed 63 characters.";
          } else if (!re.test(key[1])) {
            annotationsErrors[`annotations.key[${i}]`] =
              'Start and end with alphanumeric characters only, limiting characters in between to dashes, underscores, and alphanumerics.';
          }
        } else if (key.length === 1) {
          if (key[0].length > 63) {
            annotationsErrors[`annotations.key[${i}]`] = "Name (the segment after a slash (/), or only segment if no slash) can't exceed 63 characters.";
          } else if (!re.test(key[0])) {
            annotationsErrors[`annotations.key[${i}]`] =
              'Start and end with alphanumeric characters only, limiting characters in between to dashes, underscores, and alphanumerics.';
          }
        }
      }
      if (!a.Value) {
        annotationsErrors[`annotations.value[${i}]`] = 'Value is required.';
      }
      duplicatedAnnotations.push(a.Key);
    });
    return annotationsErrors;
  }
}

export default KubernetesAnnotationsUtils;
