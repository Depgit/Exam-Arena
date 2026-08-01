'use client';

import React, { useState } from 'react';
import { 
  Users, 
  MessageSquare, 
  UserPlus, 
  Zap, 
  Send, 
  Globe, 
  ShieldCheck, 
  Smile, 
  Check 
} from 'lucide-react';
import { Friend, ChatMessage, PlayerProfile } from '@/lib/types';
import { soundManager } from '@/lib/audio';

interface FriendsAndChatViewProps {
  friends: Friend[];
  globalChat: ChatMessage[];
  userProfile: PlayerProfile;
  onSendChatMessage: (text: string) => void;
  onAddFriend: (username: string) => void;
  onChallengeFriend: (friend: Friend) => void;
}

export const FriendsAndChatView: React.FC<FriendsAndChatViewProps> = ({
  friends,
  globalChat,
  userProfile,
  onSendChatMessage,
  onAddFriend,
  onChallengeFriend,
}) => {
  const [chatInput, setChatInput] = useState('');
  const [addFriendInput, setAddFriendInput] = useState('');
  const [showAddModal, setShowAddModal] = useState(false);
  const [addSuccessMsg, setAddSuccessMsg] = useState('');

  const handleSend = (e: React.FormEvent) => {
    e.preventDefault();
    if (!chatInput.trim()) return;
    soundManager.playClick();
    onSendChatMessage(chatInput.trim());
    setChatInput('');
  };

  const handleAddFriendSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (!addFriendInput.trim()) return;
    soundManager.playClick();
    onAddFriend(addFriendInput.trim());
    setAddSuccessMsg(`Friend request sent to ${addFriendInput.trim()}!`);
    setAddFriendInput('');
    setTimeout(() => {
      setAddSuccessMsg('');
      setShowAddModal(false);
    }, 1500);
  };

  return (
    <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 space-y-8 animate-fadeIn">
      {/* Header */}
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <div>
          <div className="inline-flex items-center gap-2 rounded-full border border-cyan-500/30 bg-cyan-500/10 px-3 py-1 text-xs font-bold text-cyan-400">
            <Users className="h-3.5 w-3.5" />
            <span>COMMUNITY & SOCIAL HUB</span>
          </div>
          <h1 className="text-2xl font-black text-white sm:text-3xl tracking-tight mt-1">
            Friends List & Global Dual Chat
          </h1>
          <p className="text-xs text-slate-400">
            Connect with other speed duelists, send 1v1 challenges, and chat live in real time.
          </p>
        </div>

        <button
          onClick={() => {
            soundManager.playClick();
            setShowAddModal(true);
          }}
          className="flex items-center gap-2 rounded-2xl bg-gradient-to-r from-cyan-500 to-blue-500 px-4 py-2.5 text-xs font-bold text-slate-950 shadow-lg hover:scale-105 transition-transform"
        >
          <UserPlus className="h-4 w-4" />
          <span>Add Friend</span>
        </button>
      </div>

      {/* Add Friend Modal */}
      {showAddModal && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-slate-950/80 p-4 backdrop-blur-sm animate-fadeIn">
          <div className="w-full max-w-md rounded-3xl border border-slate-800 bg-slate-900 p-6 shadow-2xl space-y-4">
            <h3 className="text-base font-bold text-white">Add Friend by Username</h3>
            <form onSubmit={handleAddFriendSubmit} className="space-y-3">
              <input
                type="text"
                value={addFriendInput}
                onChange={(e) => setAddFriendInput(e.target.value)}
                placeholder="Enter username (e.g. QuantumMind)"
                className="w-full rounded-2xl border border-slate-700 bg-slate-950 px-4 py-2.5 text-sm text-white focus:border-cyan-500 focus:outline-none"
              />
              {addSuccessMsg && (
                <p className="text-xs font-bold text-emerald-400 flex items-center gap-1">
                  <Check className="h-3.5 w-3.5" /> {addSuccessMsg}
                </p>
              )}
              <div className="flex justify-end gap-2 pt-2">
                <button
                  type="button"
                  onClick={() => setShowAddModal(false)}
                  className="rounded-xl border border-slate-800 px-4 py-2 text-xs font-bold text-slate-400 hover:text-white"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="rounded-xl bg-cyan-500 px-4 py-2 text-xs font-bold text-slate-950 hover:bg-cyan-400"
                >
                  Send Invite
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Two Column Layout: Friends List (Left) + Live Global Chat Room (Right) */}
      <div className="grid grid-cols-1 gap-8 lg:grid-cols-12">
        {/* Left Column: Friends List */}
        <div className="space-y-4 lg:col-span-5">
          <div className="rounded-3xl border border-slate-800 bg-slate-900/90 p-5 space-y-4 shadow-xl">
            <h2 className="text-sm font-bold text-slate-300 uppercase tracking-wider flex items-center justify-between">
              <span>Friends ({friends.length})</span>
              <span className="text-[10px] text-emerald-400">
                ● {friends.filter((f) => f.status === 'online').length} Online
              </span>
            </h2>

            <div className="space-y-3">
              {friends.map((friend) => (
                <div
                  key={friend.id}
                  className="flex items-center justify-between rounded-2xl border border-slate-800 bg-slate-950/60 p-3 hover:border-slate-700 transition-colors"
                >
                  <div className="flex items-center gap-3 min-w-0">
                    <div className="relative">
                      <img
                        src={friend.avatarUrl}
                        alt={friend.username}
                        className="h-10 w-10 rounded-xl object-cover"
                      />
                      <span
                        className={`absolute -bottom-0.5 -right-0.5 h-3 w-3 rounded-full border-2 border-slate-950 ${
                          friend.status === 'online'
                            ? 'bg-emerald-400'
                            : friend.status === 'in_match'
                            ? 'bg-orange-400'
                            : 'bg-slate-600'
                        }`}
                      />
                    </div>

                    <div className="min-w-0">
                      <p className="text-xs font-bold text-white truncate">{friend.username}</p>
                      <p className="text-[10px] text-slate-400">{friend.title} • {friend.mmr} MMR</p>
                    </div>
                  </div>

                  <button
                    onClick={() => {
                      soundManager.playMatchStart();
                      onChallengeFriend(friend);
                    }}
                    className="flex items-center gap-1.5 rounded-xl bg-gradient-to-r from-orange-500 to-amber-500 px-3 py-1.5 text-xs font-bold text-slate-950 shadow-md hover:scale-105 transition-transform"
                  >
                    <Zap className="h-3.5 w-3.5" /> 1v1 Dual
                  </button>
                </div>
              ))}
            </div>
          </div>
        </div>

        {/* Right Column: Global Dual Chat Room */}
        <div className="lg:col-span-7">
          <div className="flex h-[520px] flex-col rounded-3xl border border-slate-800 bg-slate-900/90 shadow-2xl overflow-hidden">
            {/* Chat Header */}
            <div className="flex items-center justify-between border-b border-slate-800 p-4 bg-slate-950/50">
              <div className="flex items-center gap-2">
                <Globe className="h-4 w-4 text-cyan-400" />
                <h2 className="text-sm font-bold text-white">Global Duelists Chat</h2>
              </div>
              <span className="text-[10px] text-slate-400">Live Real-Time Room</span>
            </div>

            {/* Message Feed */}
            <div className="flex-1 overflow-y-auto p-4 space-y-3">
              {globalChat.map((msg) => {
                const isSelf = msg.senderId === userProfile.id;

                return (
                  <div
                    key={msg.id}
                    className={`flex items-start gap-2.5 ${isSelf ? 'flex-row-reverse' : ''}`}
                  >
                    <img
                      src={msg.senderAvatar}
                      alt={msg.senderName}
                      className="h-8 w-8 rounded-xl object-cover mt-1"
                    />

                    <div
                      className={`max-w-[75%] rounded-2xl p-3 text-xs ${
                        isSelf
                          ? 'bg-gradient-to-r from-orange-500 to-amber-500 text-slate-950 font-semibold'
                          : 'border border-slate-800 bg-slate-950 text-slate-200'
                      }`}
                    >
                      <div className="flex items-center justify-between gap-2 mb-1">
                        <span className={`font-bold ${isSelf ? 'text-slate-950' : 'text-amber-400'}`}>
                          {msg.senderName}
                        </span>
                        <span className={`text-[9px] ${isSelf ? 'text-slate-800' : 'text-slate-500'}`}>
                          {new Date(msg.timestamp).toLocaleTimeString([], {
                            hour: '2-digit',
                            minute: '2-digit',
                          })}
                        </span>
                      </div>
                      <p className="leading-relaxed">{msg.text}</p>
                    </div>
                  </div>
                );
              })}
            </div>

            {/* Chat Input Bar */}
            <form onSubmit={handleSend} className="border-t border-slate-800 p-3 bg-slate-950/80 flex gap-2">
              <input
                type="text"
                value={chatInput}
                onChange={(e) => setChatInput(e.target.value)}
                placeholder="Type message or challenge to global room..."
                className="flex-1 rounded-2xl border border-slate-800 bg-slate-900 px-4 py-2 text-xs text-white placeholder-slate-500 focus:border-cyan-500 focus:outline-none"
              />
              <button
                type="submit"
                className="flex h-9 w-9 items-center justify-center rounded-2xl bg-cyan-500 text-slate-950 hover:bg-cyan-400 transition-transform active:scale-95"
              >
                <Send className="h-4 w-4" />
              </button>
            </form>
          </div>
        </div>
      </div>
    </div>
  );
};
