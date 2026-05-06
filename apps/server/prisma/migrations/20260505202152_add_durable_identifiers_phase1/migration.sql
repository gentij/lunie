-- AlterTable
ALTER TABLE "Workflow"
ADD COLUMN "key" TEXT,
ADD COLUMN "runSequence" INTEGER NOT NULL DEFAULT 0;

-- AlterTable
ALTER TABLE "Trigger"
ADD COLUMN "key" TEXT;

-- AlterTable
ALTER TABLE "WorkflowRun"
ADD COLUMN "number" INTEGER;
