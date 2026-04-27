import { Test, TestingModule } from '@nestjs/testing';
import { HealthService } from './health.service';
import { ConfigService } from '@nestjs/config';
import { HealthResDto } from './dto/health.dto';
import { PrismaService } from '@lunie/db-access';

describe('HealthService', () => {
  let service: HealthService;

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      providers: [
        HealthService,
        ConfigService,
        {
          provide: PrismaService,
          useValue: {
            healthInfo: jest.fn().mockResolvedValue({ ok: true, latencyMs: 1 }),
          },
        },
      ],
    }).compile();

    service = module.get<HealthService>(HealthService);
  });

  it('should be defined', () => {
    expect(service).toBeDefined();
  });

  it('should return health information', async () => {
    const result: HealthResDto = await service.health();

    expect(result).toEqual(
      expect.objectContaining({
        status: 'ok',
        version: expect.any(String) as string,
        uptime: expect.any(Number) as number,
        db: {
          latencyMs: expect.any(Number) as number,
          ok: true,
        },
      }),
    );
  });

  it('should return a positive uptime', async () => {
    const result: HealthResDto = await service.health();

    expect(result.uptime).toBeGreaterThan(0);
  });
});
