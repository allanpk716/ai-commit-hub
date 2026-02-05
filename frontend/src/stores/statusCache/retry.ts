/**
 * StatusCacheRetry - 重试逻辑
 * 负责处理失败操作的重试策略
 */

export interface RetryOptions {
  maxAttempts?: number;
  initialDelay?: number; // ms
  backoffMultiplier?: number;
}

export interface RetryResult<T> {
  success: boolean;
  data?: T;
  error?: Error;
  attempts: number;
}

/**
 * 带指数退避的重试函数
 */
export async function withRetry<T>(
  operation: () => Promise<T>,
  options: RetryOptions = {},
): Promise<RetryResult<T>> {
  const {
    maxAttempts = 3,
    initialDelay = 1000,
    backoffMultiplier = 2,
  } = options;

  let lastError: Error | undefined;
  let delay = initialDelay;

  for (let attempt = 1; attempt <= maxAttempts; attempt++) {
    try {
      const data = await operation();

      return {
        success: true,
        data,
        attempts: attempt,
      };
    } catch (error) {
      lastError = error as Error;

      // 如果是最后一次尝试，不再等待
      if (attempt < maxAttempts) {
        await sleep(delay);
        delay *= backoffMultiplier;
      }
    }
  }

  return {
    success: false,
    error: lastError,
    attempts: maxAttempts,
  };
}

/**
 * 延迟函数
 */
function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

/**
 * RetryManager - 重试管理器
 */
export class RetryManager {
  private static defaultOptions: RetryOptions = {
    maxAttempts: 3,
    initialDelay: 1000,
    backoffMultiplier: 2,
  };

  /**
   * 设置默认重试选项
   */
  static setDefaultOptions(options: RetryOptions): void {
    this.defaultOptions = { ...this.defaultOptions, ...options };
  }

  /**
   * 使用默认选项执行重试
   */
  static async execute<T>(
    operation: () => Promise<T>,
    options?: RetryOptions,
  ): Promise<RetryResult<T>> {
    return withRetry(operation, {
      ...this.defaultOptions,
      ...options,
    });
  }

  /**
   * 批量重试多个操作
   */
  static async executeBatch<T>(
    operations: Array<() => Promise<T>>,
    options?: RetryOptions,
  ): Promise<RetryResult<T>[]> {
    const results = await Promise.all(
      operations.map((op) => this.execute(op, options)),
    );

    return results;
  }

  /**
   * 获取当前默认选项
   */
  static getDefaultOptions(): Readonly<RetryOptions> {
    return { ...this.defaultOptions };
  }
}
