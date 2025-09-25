import React, { useEffect, useRef } from 'react';
import styles from './MessageList.module.css';
import { useChatStore } from '../../stores/chats';

const MessageList: React.FC = () => {
    const messagesEndRef = useRef<HTMLDivElement>(null);
    const currentChat = useChatStore((state: { currentChat: any; }) => state.currentChat);
    const messages = currentChat?.messages || [];

    useEffect(() => {
        messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
    }, [messages]);

    // useEffect(() => {
    //   console.log("Messages being rendered:", messages);
    // }, [messages]);

    if (messages.length === 0) {
        return null;
    }

    return (
        <div className={styles.messagesArea}>
            {messages.map((msg: { role: string | number; id: React.Key | null | undefined; content: string | number | bigint | boolean | React.ReactElement<unknown, string | React.JSXElementConstructor<any>> | Iterable<React.ReactNode> | React.ReactPortal | Promise<string | number | bigint | boolean | React.ReactPortal | React.ReactElement<unknown, string | React.JSXElementConstructor<any>> | Iterable<React.ReactNode> | null | undefined> | null | undefined; }) => {

                const classNames = [styles.messageBubble];
                
                if (msg.role && styles[msg.role]) {
                    classNames.push(styles[msg.role]);
                }

                return (
                    <div
                        key={msg.id}
                        className={classNames.join(' ')} 
                    >
                        <p>{msg.content}</p>
                    </div>
                );
            })}
            <div ref={messagesEndRef} />
        </div>
    );
};

export default MessageList;