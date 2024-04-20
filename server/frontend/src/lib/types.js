
/**
 * @typedef user
 * @type {object}
 * @property {number} id
 * @property {string} name
 * @property {string} user_name
 */

/**
 * @typedef message
 * @type {object}
 * @property {number} id
 * @property {number} sender_id
 * @property {number} receiver_id
 * @property {string} message_text
 */

/** 
 * @typedef inbox
 * @type {{user: user, messages: [message]}}
 * 
 */

/**
 * @typedef data
 * @type {object}
 * @property {string} data_type 
 * @property {string} error 
 * @property {string} search_term 
 * @property {user} user
 * @property {message} message
 * @property {[{user: user, messages: [message]}]} all_inboxes
 */