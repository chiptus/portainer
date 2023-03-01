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
    Object.keys(annotations).forEach((k) => {
      if (!KubernetesSystem_AnnotationsToSkip[k] && k !== KubernetesPortainerApplicationNote) {
        const annotation = new KubernetesFormValueAnnotation();
        annotation.Key = k;
        annotation.Value = annotations[k];
        annotation.ID = uuidv4();

        res.push(annotation);
      }
    });
    return res;
  }

  static formValuesToKubeAnnotations(formValues) {
    const annotations = {};
    if (!formValues.Annotations) {
      return annotations;
    }
    formValues.Annotations.forEach((a) => {
      annotations[a.Key] = a.Value;
    });
    return annotations;
  }

  static validateAnnotations(annotations) {
    const duplicatedAnnotations = [];
    const annotationsErrors = {};
    annotations.forEach((a, i) => {
      if (!a.Key) {
        annotationsErrors[`annotations.key[${i}]`] = 'Annotation key is required';
      } else if (duplicatedAnnotations.includes(a.Key)) {
        annotationsErrors[`annotations.key[${i}]`] = 'Annotation cannot be duplicated';
      }
      if (!a.Value) {
        annotationsErrors[`annotations.value[${i}]`] = 'Annotation value is required';
      }
      duplicatedAnnotations.push(a.Key);
    });
    return annotationsErrors;
  }
}

export default KubernetesAnnotationsUtils;
