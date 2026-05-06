import { Test, TestingModule } from '@nestjs/testing';
import { StepRunController } from './step-run.controller';
import { StepRunService } from './step-run.service';
import { createStepRunFixture } from 'test/step-run/step-run.fixtures';

describe('StepRunController', () => {
  let controller: StepRunController;
  let service: StepRunService;

  beforeEach(async () => {
    const module: TestingModule = await Test.createTestingModule({
      controllers: [StepRunController],
      providers: [
        {
          provide: StepRunService,
          useValue: {
            list: jest.fn(),
            get: jest.fn(),
          },
        },
      ],
    }).compile();

    controller = module.get<StepRunController>(StepRunController);
    service = module.get<StepRunService>(StepRunService);
  });

  it('list() calls StepRunService.list()', async () => {
    const list = [createStepRunFixture({ id: 'sr_1' })];
    const listSpy = jest.spyOn(service, 'list').mockResolvedValue({
      items: list,
      pagination: {
        page: 1,
        pageSize: 25,
        total: 1,
        totalPages: 1,
        hasNext: false,
        hasPrev: false,
      },
    });

    await expect(
      controller.list('wf_1', 'wfr_1', { page: 1, pageSize: 25 }),
    ).resolves.toEqual({
      items: list,
      pagination: {
        page: 1,
        pageSize: 25,
        total: 1,
        totalPages: 1,
        hasNext: false,
        hasPrev: false,
      },
    });
    expect(listSpy).toHaveBeenCalledWith({
      workflowId: 'wf_1',
      workflowRunId: 'wfr_1',
      page: 1,
      pageSize: 25,
    });
  });

  it('get() calls StepRunService.get()', async () => {
    const step = createStepRunFixture({ id: 'sr_1', workflowRunId: 'wfr_1' });
    const getSpy = jest.spyOn(service, 'get').mockResolvedValue(step);

    await expect(controller.get('wf_1', 'wfr_1', 'sr_1')).resolves.toBe(step);
    expect(getSpy).toHaveBeenCalledWith('wf_1', 'wfr_1', 'sr_1');
  });
});
