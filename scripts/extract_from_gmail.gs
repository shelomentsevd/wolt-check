/**
 * Chunked attachment downloader from a single sender,
 * skipping files already saved, and only fetching threads
 * that actually have attachments.
 */
function saveAttachmentsFromGooseChunked() {
  const senderEmail = 'info@wolt.com';
  const batchSize   = 100;                      // bump up per‐run size
  const propKey     = 'goose_last_offset';
  const handlerName = 'saveAttachmentsFromGooseChunked';

  // 1) resume previous offset
  const props = PropertiesService.getScriptProperties();
  let offset = Number(props.getProperty(propKey)) || 0;

  // 2) prepare Drive folder
  const folderName = 'Attachments_from_' + senderEmail.replace(/[@.]/g, '_');
  let folderIter = DriveApp.getFoldersByName(folderName);
  const folder = folderIter.hasNext()
    ? folderIter.next()
    : DriveApp.createFolder(folderName);

  // 3) only grab threads that have attachments
  const query   = `from:${senderEmail} has:attachment`;
  const threads = GmailApp.search(query, offset, batchSize);

  if (threads.length === 0) {
    // all done
    props.deleteProperty(propKey);
    cleanupTrigger_(handlerName);
    Logger.log('✅ Finished – no more threads.');
    return;
  }

  // 4) process this batch
  let saved = 0, skipped = 0;
  threads.forEach(thread => {
    thread.getMessages().forEach(msg => {
      msg.getAttachments().forEach(att => {
        const name = att.getName();
        if (folder.getFilesByName(name).hasNext()) {
          skipped++;
        } else {
          folder.createFile(att.copyBlob());
          saved++;
        }
      });
    });
  });
  Logger.log(`Threads ${offset + 1}–${offset + threads.length}: saved ${saved}, skipped ${skipped}`);

  // 5) advance & persist offset
  offset += threads.length;
  props.setProperty(propKey, offset);

  // 6) schedule next run in 1 minute
  cleanupTrigger_(handlerName);
  ScriptApp.newTrigger(handlerName)
    .timeBased()
    .after(1 * 60 * 1000)   // 60 000 ms = 1 minute
    .create();
}

/** Remove any existing trigger for the given handler */
function cleanupTrigger_(handlerName) {
  ScriptApp.getProjectTriggers()
    .filter(t => t.getHandlerFunction() === handlerName)
    .forEach(t => ScriptApp.deleteTrigger(t));
}

