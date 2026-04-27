import { Global, Module } from '@nestjs/common';
import { PrismaService } from '@lunie/db-access';

@Global()
@Module({
  providers: [PrismaService],
  exports: [PrismaService],
})
export class PrismaModule {}
