import { Module } from '@nestjs/common';
import { PrismaModule } from 'src/prisma/prisma.module';
import { CryptoModule } from 'src/crypto/crypto.module';
import { SecretController } from './secret.controller';
import { SecretRepository } from '@lunie/db-access';
import { SecretService } from './secret.service';

@Module({
  imports: [PrismaModule, CryptoModule],
  controllers: [SecretController],
  providers: [SecretService, SecretRepository],
  exports: [SecretService],
})
export class SecretModule {}
