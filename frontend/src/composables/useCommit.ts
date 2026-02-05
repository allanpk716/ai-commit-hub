import { ref, computed, onUnmounted } from "vue";
import { useStatusCache } from "@/stores/statusCache";
import { APP_EVENTS } from "@/constants/events";
import { CommitProject, GenerateCommit } from "@/wailsjs/go/main/App";
import { EventsOn, EventsOff } from "@/wailsjs/runtime/runtime";

export function useCommit() {
  const statusCache = useStatusCache();
  const isGenerating = ref(false);
  const generatedMessage = ref("");
  const commitError = ref("");

  const canCommit = computed(
    () => !isGenerating.value && generatedMessage.value.trim().length > 0,
  );

  // 监听流式输出事件
  const handleCommitDelta = (delta: string) => {
    generatedMessage.value += delta;
  };

  // 监听完成事件
  const handleCommitComplete = (data: { success: boolean; error?: string }) => {
    isGenerating.value = false;
    if (!data.success) {
      commitError.value = data.error || "生成失败";
    }
  };

  // 注册事件监听
  EventsOn(APP_EVENTS.COMMIT_DELTA, handleCommitDelta);
  EventsOn(APP_EVENTS.COMMIT_COMPLETE, handleCommitComplete);

  // 清理事件监听
  onUnmounted(() => {
    EventsOff(APP_EVENTS.COMMIT_DELTA, handleCommitDelta);
    EventsOff(APP_EVENTS.COMMIT_COMPLETE, handleCommitComplete);
  });

  /**
   * 生成 commit 消息
   */
  async function generateMessage(projectPath: string) {
    isGenerating.value = true;
    generatedMessage.value = "";
    commitError.value = "";

    try {
      await GenerateCommit(projectPath);
    } catch (error: unknown) {
      isGenerating.value = false;
      commitError.value = error instanceof Error ? error.message : "生成失败";
    }
  }

  /**
   * 提交 commit
   */
  async function commit(projectPath: string, message: string) {
    const rollback = statusCache.updateOptimistic(projectPath, {
      hasUncommittedChanges: false,
    });

    try {
      await CommitProject(projectPath, message);
      await statusCache.refresh(projectPath, { force: true });
    } catch (error: unknown) {
      rollback?.();
      throw error;
    }
  }

  /**
   * 清空生成的消息
   */
  function clearMessage() {
    generatedMessage.value = "";
    commitError.value = "";
  }

  return {
    isGenerating,
    generatedMessage,
    commitError,
    canCommit,
    generateMessage,
    commit,
    clearMessage,
  };
}
