import { useEffect, useState } from 'react';

import { validationSchema } from './EnvironmentNameForm/validation';

export function useIsKaasNameValid(environmentName: string) {
  const [isNameValid, setisNameValid] = useState(false);

  useEffect(() => {
    async function validateName() {
      const nameValidation = validationSchema();
      const nameValid = await nameValidation.isValid({
        name: environmentName,
      });
      setisNameValid(nameValid);
    }
    validateName();
  }, [environmentName]);
  return isNameValid;
}
