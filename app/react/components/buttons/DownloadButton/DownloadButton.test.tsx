import { render, fireEvent } from '@testing-library/react';
import '@testing-library/jest-dom';
import saveAs from 'file-saver';

import { DownloadButton } from './DownloadButton';

jest.mock('file-saver', () => ({
  saveAs: jest.fn(),
}));

jest.mock('@/portainer/services/notifications', () => ({
  notifyError: jest.fn(),
}));

describe('DownloadButton', () => {
  beforeEach(() => {
    jest.clearAllMocks();
  });

  it('should display button with children', () => {
    const fileName = 'test.txt';
    const fileContent = 'test content';
    const { getByText } = render(
      <DownloadButton fileName={fileName} fileContent={fileContent}>
        Download
      </DownloadButton>
    );
    const button = getByText('Download');

    expect(button).toBeTruthy();
  });

  it('should show disabled when fileContent is an empty string', () => {
    const fileName = 'test.txt';
    const fileContent = '';
    const { getByText } = render(
      <DownloadButton fileName={fileName} fileContent={fileContent}>
        Download
      </DownloadButton>
    );
    const button = getByText('Download');

    expect(button).toBeDisabled();
  });

  it('should show disabled when fileName is an empty string', () => {
    const fileName = '';
    const fileContent = 'test content';
    const { getByText } = render(
      <DownloadButton fileName={fileName} fileContent={fileContent}>
        Download
      </DownloadButton>
    );
    const button = getByText('Download');

    expect(button).toBeDisabled();
  });

  it('should call saveAs when fileContent and fileName are provided', () => {
    const mockSaveAs = jest.fn() as unknown as typeof saveAs;
    // Arrange
    const fileName = 'test.txt';
    const fileContent = 'test content';
    const { getByText } = render(
      <DownloadButton
        fileName={fileName}
        fileContent={fileContent}
        saveFunction={mockSaveAs}
      >
        Download
      </DownloadButton>
    );
    const button = getByText('Download');

    fireEvent.click(button);

    expect(mockSaveAs).toHaveBeenCalled();
  });
});
