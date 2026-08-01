'use client';

import React from 'react';
import { 
  X, 
  Bell, 
  Swords, 
  Trophy, 
  Check, 
  Trash2, 
  Sparkles, 
  UserPlus, 
  Zap 
} from 'lucide-react';
import { AppNotification } from '@/lib/types';
import { notificationManager } from '@/lib/notification';
import { soundManager } from '@/lib/audio';

interface NotificationCenterProps {
  isOpen: boolean;
  onClose: () => void;
  notifications: AppNotification[];
  onClearAll: () => void;
  onMarkAllRead: () => void;
  onAcceptChallenge: (roomCode: string) => void;
}

export const NotificationCenter: React.FC<NotificationCenterProps> = ({
  isOpen,
  onClose,
  notifications,
  onClearAll,
  onMarkAllRead,
  onAcceptChallenge,
}) => {
  const [pushPermission, setPushPermission] = React.useState<NotificationPermission>(
    notificationManager.getPermissionStatus()
  );

  if (!isOpen) return null;

  const handleRequestPush = async () => {
    soundManager.playClick();
    const granted = await notificationManager.requestPermission();
    setPushPermission(notificationManager.getPermissionStatus());
    if (granted) {
      notificationManager.sendNotification('⚡ Push Notifications Active!', {
        body: 'You will now receive instant turn alerts and 1v1 challenge invites!',
      });
    }
  };

  const getIcon = (type: AppNotification['type']) => {
    switch (type) {
      case 'challenge':
        return <Swords className="h-5 w-5 text-orange-400" />;
      case 'turn_update':
        return <Zap className="h-5 w-5 text-amber-400" />;
      case 'achievement':
        return <Trophy className="h-5 w-5 text-emerald-400" />;
      case 'friend_request':
        return <UserPlus className="h-5 w-5 text-cyan-400" />;
      case 'match_result':
        return <Sparkles className="h-5 w-5 text-purple-400" />;
      default:
        return <Bell className="h-5 w-5 text-slate-400" />;
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex justify-end bg-slate-950/60 backdrop-blur-sm animate-fadeIn">
      <div className="flex h-full w-full max-w-md flex-col border-l border-slate-800 bg-slate-900 text-white shadow-2xl">
        {/* Drawer Header */}
        <div className="flex items-center justify-between border-b border-slate-800 p-4">
          <div className="flex items-center gap-2">
            <Bell className="h-5 w-5 text-orange-400" />
            <h2 className="text-lg font-bold">Notifications & Alerts</h2>
          </div>
          <button
            onClick={() => {
              soundManager.playClick();
              onClose();
            }}
            className="rounded-lg p-1 text-slate-400 transition-colors hover:bg-slate-800 hover:text-white"
          >
            <X className="h-5 w-5" />
          </button>
        </div>

        {/* Browser Push Permission Banner */}
        <div className="border-b border-slate-800 bg-slate-950/50 p-4">
          <div className="flex items-center justify-between gap-3">
            <div>
              <p className="text-xs font-semibold text-slate-200">Browser Push Notifications</p>
              <p className="text-[11px] text-slate-400">
                Get notified when opponent answers or challenges arrive while app is inactive.
              </p>
            </div>
            {pushPermission === 'granted' ? (
              <span className="flex items-center gap-1 text-[11px] font-bold text-emerald-400">
                <Check className="h-3.5 w-3.5" /> Enabled
              </span>
            ) : (
              <button
                onClick={handleRequestPush}
                className="rounded-xl bg-orange-500 px-3 py-1.5 text-xs font-bold text-slate-950 transition-transform hover:scale-105 active:scale-95"
              >
                Enable
              </button>
            )}
          </div>
        </div>

        {/* Action controls */}
        <div className="flex items-center justify-between border-b border-slate-800 px-4 py-2 bg-slate-900/50">
          <button
            onClick={() => {
              soundManager.playClick();
              onMarkAllRead();
            }}
            className="text-xs text-amber-400 hover:underline"
          >
            Mark all read
          </button>
          <button
            onClick={() => {
              soundManager.playClick();
              onClearAll();
            }}
            className="flex items-center gap-1 text-xs text-rose-400 hover:underline"
          >
            <Trash2 className="h-3.5 w-3.5" /> Clear all
          </button>
        </div>

        {/* Notification List */}
        <div className="flex-1 overflow-y-auto p-4 space-y-3">
          {notifications.length === 0 ? (
            <div className="flex h-64 flex-col items-center justify-center text-center">
              <Bell className="h-10 w-10 text-slate-600 mb-2" />
              <p className="text-sm font-medium text-slate-400">No notifications yet</p>
              <p className="text-xs text-slate-500 mt-1">
                You will be alerted when turn updates or challenges arrive!
              </p>
            </div>
          ) : (
            notifications.map((notif) => (
              <div
                key={notif.id}
                className={`flex gap-3 rounded-2xl border p-3.5 transition-all ${
                  notif.read
                    ? 'border-slate-800 bg-slate-950/40 text-slate-300'
                    : 'border-orange-500/30 bg-gradient-to-r from-orange-500/10 via-slate-900 to-slate-900 text-white ring-1 ring-orange-500/20'
                }`}
              >
                <div className="mt-0.5 flex h-9 w-9 shrink-0 items-center justify-center rounded-xl bg-slate-800">
                  {getIcon(notif.type)}
                </div>
                <div className="flex-1 min-w-0">
                  <div className="flex items-center justify-between gap-2">
                    <h4 className="text-xs font-bold truncate text-amber-300">{notif.title}</h4>
                    <span className="text-[10px] text-slate-500 shrink-0">
                      {new Date(notif.timestamp).toLocaleTimeString([], {
                        hour: '2-digit',
                        minute: '2-digit',
                      })}
                    </span>
                  </div>
                  <p className="text-xs text-slate-300 mt-1 leading-relaxed">{notif.message}</p>

                  {/* Challenge Action Button */}
                  {notif.type === 'challenge' && notif.actionPayload?.roomCode && (
                    <button
                      onClick={() => {
                        soundManager.playMatchStart();
                        onAcceptChallenge(notif.actionPayload!.roomCode!);
                        onClose();
                      }}
                      className="mt-2.5 flex items-center gap-1.5 rounded-xl bg-gradient-to-r from-orange-500 to-amber-500 px-3 py-1.5 text-xs font-bold text-slate-950 shadow-md transition-transform hover:scale-105"
                    >
                      <Swords className="h-3.5 w-3.5" /> Accept & Enter Room
                    </button>
                  )}
                </div>
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  );
};
