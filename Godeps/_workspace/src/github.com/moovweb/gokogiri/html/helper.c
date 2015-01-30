#include "helper.h"
#include "../xml/helper.h"
#include <string.h>

htmlDocPtr htmlParse(void *buffer, int buffer_len, void *url, void *encoding, int options, void *error_buffer, int error_buffer_len) {
	const char *c_buffer       = (char*)buffer;
	const char *c_url          = (char*)url;
	const char *c_encoding     = (char*)encoding;
	xmlDoc *doc = NULL;

	xmlResetLastError();
	doc = htmlReadMemory(c_buffer, buffer_len, c_url, c_encoding, options);

	return doc;
}

xmlNode* htmlParseFragment(void *doc, void *buffer, int buffer_len, void *url, int options, void *error_buffer, int error_buffer_len) {
	xmlNode* root_element = NULL;
	xmlParserErrors errCode;
	errCode = xmlParseInNodeContext((xmlNodePtr)doc, buffer, buffer_len, options, &root_element);
	if (errCode != XML_ERR_OK) {
		return NULL;
	}
	return root_element;
}

xmlNode* htmlParseFragmentAsDoc(void *doc, void *buffer, int buffer_len, void *url, void *encoding, int options, void *error_buffer, int error_buffer_len) {
	xmlDoc* tmpDoc = NULL;
	xmlNode* tmpRoot = NULL;
	tmpDoc = htmlReadMemory((char*)buffer, buffer_len, (char*)url, (char*)encoding, options);
	if (tmpDoc == NULL) {
		return NULL;
	}
	tmpRoot = xmlDocGetRootElement(tmpDoc);
	if (tmpRoot == NULL) {
		return NULL;
	}
	tmpRoot = xmlDocCopyNode(tmpRoot, doc, 1);
	xmlFreeDoc(tmpDoc);
	return tmpRoot;
}
