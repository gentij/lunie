import { Test } from '@nestjs/testing';
import { CronTriggerScheduler } from './cron-trigger.scheduler';
import { PrismaService } from '@lunie/db-access';
import { OrchestrationService } from 'src/core/orchestration.service';

describe('CronTriggerScheduler', () => {
  it('does not start workflow when not due', async () => {
    const prisma = {
      trigger: {
        findMany: jest.fn().mockResolvedValue([
          {
            id: 'tr_1',
            type: 'CRON',
            isActive: true,
            config: { cron: '0 0 1 1 *', timezone: 'UTC', input: {} },
            workflow: { id: 'wf_1', latestVersionId: 'wfv_1' },
          },
        ]),
      },
    };

    const orchestration = {
      startWorkflow: jest.fn(),
    };

    const moduleRef = await Test.createTestingModule({
      providers: [
        CronTriggerScheduler,
        { provide: PrismaService, useValue: prisma },
        { provide: OrchestrationService, useValue: orchestration },
      ],
    }).compile();

    const svc = moduleRef.get(CronTriggerScheduler);
    await svc.tick();

    expect(orchestration.startWorkflow).not.toHaveBeenCalled();
  });
});
