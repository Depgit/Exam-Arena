// Web Notifications API & Local Notification Helper
// Handles browser push notifications, turn updates, and match alerts

class NotificationEngine {
  private permission: NotificationPermission = 'default';

  constructor() {
    if (typeof window !== 'undefined' && 'Notification' in window) {
      this.permission = Notification.permission;
    }
  }

  public getPermissionStatus(): NotificationPermission {
    if (typeof window !== 'undefined' && 'Notification' in window) {
      this.permission = Notification.permission;
    }
    return this.permission;
  }

  public async requestPermission(): Promise<boolean> {
    if (typeof window === 'undefined' || !('Notification' in window)) {
      console.warn('Web Notifications are not supported in this browser.');
      return false;
    }

    try {
      const result = await Notification.requestPermission();
      this.permission = result;
      return result === 'granted';
    } catch (e) {
      console.error('Error requesting notification permission:', e);
      return false;
    }
  }

  public sendNotification(title: string, options?: NotificationOptions) {
    if (typeof window === 'undefined' || !('Notification' in window)) {
      return;
    }

    if (this.permission === 'granted') {
      try {
        const notif = new Notification(title, {
          icon: '/favicon.ico',
          badge: '/favicon.ico',
          ...options,
        });

        notif.onclick = () => {
          window.focus();
          notif.close();
        };
      } catch (err) {
        console.error('Failed to trigger notification:', err);
      }
    }
  }

  public notifyTurnUpdate(opponentName: string, roundNumber: number) {
    this.sendNotification(`⚡ Dual Exam: Your Turn!`, {
      body: `${opponentName} has submitted their answer for Round ${roundNumber}! Hurry and lock in your answer.`,
      tag: 'turn_update',
    });
  }

  public notifyChallengeReceived(challengerName: string, category: string) {
    this.sendNotification(`⚔️ 1v1 Dual Challenge Invited!`, {
      body: `${challengerName} challenged you to a 1v1 ${category} Dual Exam! Tap to enter room.`,
      tag: 'challenge_invite',
    });
  }

  public notifyMatchResult(winnerName: string, isWinner: boolean, scoreText: string) {
    this.sendNotification(isWinner ? `🏆 VICTORY in Dual Exam!` : `💔 Match Complete`, {
      body: isWinner 
        ? `You won against ${winnerName}! Score: ${scoreText}`
        : `${winnerName} won this round. Final Score: ${scoreText}`,
      tag: 'match_result',
    });
  }
}

export const notificationManager = new NotificationEngine();
