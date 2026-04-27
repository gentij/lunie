import { Test } from '@nestjs/testing';
import { AuthBootstrapService } from './auth-bootstrap.service';
import { ApiTokenService } from 'src/api-token/api-token.service';
import { CryptoService } from 'src/crypto/crypto.service';
import { PinoLogger } from 'nestjs-pino';
import { ConfigService } from '@nestjs/config';

import {
  createApiTokenServiceMock,
  ApiTokenServiceMock,
} from 'test/api-token/api-token.service.mock';
import {
  createCryptoServiceMock,
  CryptoServiceMock,
} from 'test/crypto/crypto.service.mock';
import {
  createPinoLoggerMock,
  PinoLoggerMock,
} from 'test/logger/logger.service.mock';

describe('AuthBootstrapService', () => {
  let service: AuthBootstrapService;
  let apiTokenServiceMock: ApiTokenServiceMock;
  let cryptoServiceMock: CryptoServiceMock;
  let loggerMock: PinoLoggerMock;

  beforeEach(async () => {
    apiTokenServiceMock = createApiTokenServiceMock();
    cryptoServiceMock = createCryptoServiceMock();
    loggerMock = createPinoLoggerMock();

    const moduleRef = await Test.createTestingModule({
      providers: [
        AuthBootstrapService,
        { provide: ApiTokenService, useValue: apiTokenServiceMock },
        { provide: CryptoService, useValue: cryptoServiceMock },
        { provide: PinoLogger, useValue: loggerMock },
        {
          provide: ConfigService,
          useValue: {
            getOrThrow: jest.fn((key: string) => process.env[key]),
          },
        },
      ],
    }).compile();

    service = moduleRef.get(AuthBootstrapService);
  });

  afterEach(() => {
    delete process.env.LUNIE_ADMIN_TOKEN;
  });

  it('skips bootstrap if an active token already exists', async () => {
    apiTokenServiceMock.hasAnyActiveToken.mockResolvedValue(true);

    await service.onModuleInit();

    expect(apiTokenServiceMock.createAdminToken).not.toHaveBeenCalled();
    expect(cryptoServiceMock.hashApiToken).not.toHaveBeenCalled();

    expect(loggerMock.info).toHaveBeenCalled();
    expect(loggerMock.warn).not.toHaveBeenCalled();
  });

  it('creates initial admin token if none exists', async () => {
    apiTokenServiceMock.hasAnyActiveToken.mockResolvedValue(false);
    cryptoServiceMock.hashApiToken.mockReturnValue('hashed_token_abc');

    process.env.LUNIE_ADMIN_TOKEN = 'lunie_raw_token_123';

    await service.onModuleInit();

    expect(cryptoServiceMock.hashApiToken).toHaveBeenCalledWith(
      'lunie_raw_token_123',
    );
    expect(apiTokenServiceMock.createAdminToken).toHaveBeenCalledWith({
      name: 'initial-admin',
      tokenHash: 'hashed_token_abc',
    });

    const infoCalls = loggerMock.info.mock.calls.flat().join(' ');
    expect(infoCalls).toContain('initialized from environment');
  });
});
