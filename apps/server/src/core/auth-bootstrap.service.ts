import { Injectable, OnModuleInit } from '@nestjs/common';
import { ConfigService } from '@nestjs/config';
import { ApiTokenService } from '../api-token/api-token.service';
import { CryptoService } from 'src/crypto/crypto.service';
import { PinoLogger } from 'nestjs-pino';
import { Env } from 'src/config/env';

@Injectable()
export class AuthBootstrapService implements OnModuleInit {
  constructor(
    private readonly apiTokenService: ApiTokenService,
    private readonly cryptoService: CryptoService,
    private readonly logger: PinoLogger,
    private readonly configService: ConfigService<Env>,
  ) {
    this.logger.setContext(AuthBootstrapService.name);
  }

  async onModuleInit() {
    const hasToken = await this.apiTokenService.hasAnyActiveToken();

    if (hasToken) {
      this.logger.info('API token already exists, skipping bootstrap');
      return;
    }

    const rawToken: string = this.configService.getOrThrow('LUNIE_ADMIN_TOKEN');
    const tokenHash = this.cryptoService.hashApiToken(rawToken);

    await this.apiTokenService.createAdminToken({
      name: 'initial-admin',
      tokenHash,
    });

    this.logger.info('Admin API token initialized from environment');
  }
}
