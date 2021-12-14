import { AxiosError, AxiosResponse } from 'axios';
import mockAxios from 'jest-mock-axios';

import { getLicenses } from './license.service';
import type { License } from './types';

afterEach(() => {
  mockAxios.reset();
});

describe('getLicenses', () => {
  it('on success should return the server body', async () => {
    const catchFn = jest.fn();
    const thenFn = jest.fn();

    const promise = getLicenses();

    const data: License[] = [];

    mockAxios.mockResponse({ data });
    await promise.then(thenFn).catch(catchFn);

    expect(catchFn).not.toHaveBeenCalled();
    expect(thenFn).toHaveBeenCalledWith(data);
  });

  it('on failure should return the server message', async () => {
    const catchFn = jest.fn();
    const thenFn = jest.fn();

    const promise = getLicenses();

    const message = 'message';
    const details = 'details';

    mockAxios.mockError(buildAxiosError(message, details));
    await promise.then(thenFn, catchFn);

    expect(catchFn).toHaveBeenCalledWith(new Error(message));
    expect(thenFn).not.toHaveBeenCalled();
  });
});

function buildAxiosError(
  message: string,
  details: string
): Partial<AxiosError> {
  const data = { message, details };
  const response: Partial<AxiosResponse> = { data };

  return {
    response: response as AxiosResponse,
  };
}
