import { Injectable } from "@angular/core";
import { MatSnackBar, MatSnackBarConfig } from "@angular/material/snack-bar";

export type RunOpts<T = unknown> = {
    start?: string;
    success?: string;
    softSuccess?: string;
    error?: string;
    durationMs?: number;
    optimisticOnNetworkError?: boolean;
    onSuccess?: (result: T) => void;
    onFinally?: () => void;
}

@Injectable({ providedIn: "root" })
export class FeedbackService {
    constructor(private snack: MatSnackBar) {}
    
    private open(msg: string, action = 'OK', cfg?: Partial<MatSnackBarConfig>) {
        this.snack.open(msg, action, { duration: 3500, ...cfg });
    }

    info(msg: string, cfg?: Partial<MatSnackBarConfig>) {
        this.open(msg, 'OK', { panelClass: ['snack-info'], ...cfg });
    }
    success(msg: string, cfg?: Partial<MatSnackBarConfig>) {
        this.open(msg, 'OK', { panelClass: ['snack-success'], ...cfg });
    }
    error(msg: string, cfg?: Partial<MatSnackBarConfig>) {
        this.open(msg, 'Dismiss', { duration: 4500, panelClass: ['snack-error'], ...cfg });
    }
    
    /**
     * Run an async with snack feedback + optional optimistic handling.
     * `setBusy(true/false)` toggles your existing spinner flag.
     */
    async run<T>(
        action: () => Promise<T>,
        setBusy: (v: boolean) => void,
        opts: RunOpts<T> = {}
    ): Promise<void> {
        const {
            start,
            success,
            softSuccess,
            error,
            durationMs = 2500,
            optimisticOnNetworkError = false,
            onSuccess,
            onFinally
        } = opts;

        try {
            setBusy(true);
            if (start) this.info(start, { duration: durationMs });
            
            const res = await action();
            if (success) this.success(success, { duration: durationMs });
            onSuccess?.(res);
        } catch (e) {
            // If the backend restarts and the request drops, treat as soft-success
            if (optimisticOnNetworkError && softSuccess) {
                this.success(softSuccess, { duration: durationMs });
            } else if (error) {
                this.error(error);
            }
        } finally {
            setBusy(false);
            onFinally?.();
        }
    }
}