-- AlterTable
ALTER TABLE "Workflow"
ALTER COLUMN "key" SET NOT NULL;

-- AlterTable
ALTER TABLE "Trigger"
ALTER COLUMN "key" SET NOT NULL;

-- AlterTable
ALTER TABLE "WorkflowRun"
ALTER COLUMN "number" SET NOT NULL;

-- CreateIndex
CREATE UNIQUE INDEX "Workflow_key_key" ON "Workflow"("key");

-- CreateIndex
CREATE UNIQUE INDEX "Trigger_workflowId_key_key" ON "Trigger"("workflowId", "key");

-- CreateIndex
CREATE UNIQUE INDEX "WorkflowRun_workflowId_number_key" ON "WorkflowRun"("workflowId", "number");
