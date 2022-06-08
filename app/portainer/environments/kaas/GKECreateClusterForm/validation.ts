import { object, number, SchemaOf, string } from 'yup';

import { CreateGKEClusterFormValues, isGKEKaasInfo, KaasInfo } from '../types';

// for E2 machines, min is 0.5GB per vCPU
export function minGKERam(cpu: number) {
  return cpu * 0.5;
}

// for E2 machines, max is 8GB per vCPU with a 128GB maximum
export function maxGKERam(cpu: number) {
  return Math.min(cpu * 8, 128);
}

export function validationSchema(
  kaasInfo?: KaasInfo | null
): SchemaOf<CreateGKEClusterFormValues> {
  let cpuSchema = number();
  let ramSchema = number();
  let hddSchema = number();

  if (kaasInfo && isGKEKaasInfo(kaasInfo)) {
    cpuSchema = cpuSchema
      .min(
        kaasInfo.cpu.min,
        `Node vCPUs must be greater than or equal to ${kaasInfo.cpu.min}.`
      )
      .max(
        kaasInfo.cpu.max,
        `Node vCPUs must be less than or equal to ${kaasInfo.cpu.max}.`
      );
    ramSchema = ramSchema.when('cpu', (cpu, schema) =>
      schema
        .min(
          minGKERam(cpu),
          `Node RAM must be greater than or equal to ${minGKERam(cpu)} GB.`
        )
        .max(
          maxGKERam(cpu),
          `Node RAM must be less than or equal to ${maxGKERam(cpu)} GB.`
        )
    );
    hddSchema = hddSchema
      .min(
        kaasInfo.hdd.min,
        `Node disk space must be greater than or equal to ${kaasInfo.hdd.min} GB.`
      )
      .max(
        kaasInfo.hdd.max,
        `Node disk space must be less than or equal to ${kaasInfo.hdd.max} GB.`
      );
  }

  return object().shape({
    nodeCount: number()
      .integer('Node count must be a whole number.')
      .min(1, 'Node count must be greater than or equal to 1.')
      .max(1000, 'Node count must be less than or equal to 1000.')
      .required('Node count is required.'),
    cpu: cpuSchema
      .integer('CPUs must be an integer.')
      .test('2 increments', 'Number of vCPUs must a multiple of 2.', (number) =>
        number ? number % 2 === 0 : false
      )
      .required('CPUs are required.'),
    ram: ramSchema
      .required('RAM is required.')
      .test('0.25 increments', 'RAM must be in 0.25GB increments.', (number) =>
        number ? number % 0.25 === 0 : false
      ),
    hdd: hddSchema
      .integer('HDD must be an integer.')
      .required('HDD is required.'),
    networkId: string().required('Network ID is required.'),
  });
}
