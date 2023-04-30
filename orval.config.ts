import { Options } from 'orval';

const options: Record<string, Options> = {
  api: {
    input: './api/docs/swagger.yaml',
    output: {
      target: './app/generated-types/api/',
      prettier: true,
      mode: 'tags-split',
      // mock: true,
      client: 'react-query',
    },
    hooks: {
      afterAllFilesWrite: ['prettier --write', 'eslint --fix'],
    },
  },
  docker: {
    input: {
      target: 'https://docs.docker.com/engine/api/v1.41.yaml',
      override: {
        // transformer: './orval/transformer.js',
      },
    },
    output: {
      target: './app/generated-types/docker/',
      prettier: true,
      mode: 'tags-split',
      client: 'react-query',
      // mock: true,
    },
    hooks: {
      afterAllFilesWrite: ['prettier --write', 'eslint --fix'],
    },
  },
};

export default options;
