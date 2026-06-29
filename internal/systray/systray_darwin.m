#import <Cocoa/Cocoa.h>
#include "systray.h"

#if __MAC_OS_X_VERSION_MIN_REQUIRED < 101400
    #ifndef NSControlStateValueOff
      #define NSControlStateValueOff NSOffState
    #endif
    #ifndef NSControlStateValueOn
      #define NSControlStateValueOn NSOnState
    #endif
#endif

@interface SystrayMenuItem : NSObject
{
  @public
    NSNumber* menuId;
    NSNumber* parentMenuId;
    NSString* title;
    NSString* tooltip;
    short disabled;
    short checked;
}
-(id) initWithId: (int)theMenuId
withParentMenuId: (int)theParentMenuId
       withTitle: (const char*)theTitle
     withTooltip: (const char*)theTooltip
    withDisabled: (short)theDisabled
     withChecked: (short)theChecked;
@end

@implementation SystrayMenuItem
-(id) initWithId: (int)theMenuId
withParentMenuId: (int)theParentMenuId
       withTitle: (const char*)theTitle
     withTooltip: (const char*)theTooltip
    withDisabled: (short)theDisabled
     withChecked: (short)theChecked
{
  menuId = [NSNumber numberWithInt:theMenuId];
  parentMenuId = [NSNumber numberWithInt:theParentMenuId];
  title = [[NSString alloc] initWithCString:theTitle encoding:NSUTF8StringEncoding];
  tooltip = [[NSString alloc] initWithCString:theTooltip encoding:NSUTF8StringEncoding];
  disabled = theDisabled;
  checked = theChecked;
  return self;
}
@end

static NSStatusItem *statusItem = nil;
static NSMenu *systrayMenu = nil;

static NSMenuItem *find_menu_item(NSMenu *ourMenu, NSNumber *menuId) {
  NSMenuItem *foundItem = [ourMenu itemWithTag:[menuId integerValue]];
  if (foundItem != NULL) {
    return foundItem;
  }
  NSArray *items = ourMenu.itemArray;
  for (NSUInteger i = 0; i < [items count]; i++) {
    NSMenuItem *item = [items objectAtIndex:i];
    if (item.hasSubmenu) {
      foundItem = find_menu_item(item.submenu, menuId);
      if (foundItem != NULL) {
        return foundItem;
      }
    }
  }
  return NULL;
}

@interface SystrayClickHandler : NSObject
-(IBAction)menuHandler:(id)sender;
@end

@implementation SystrayClickHandler
-(IBAction)menuHandler:(id)sender {
  NSNumber* menuId = [sender representedObject];
  systray_menu_item_selected(menuId.intValue);
}
@end

static SystrayClickHandler *clickHandler = nil;

static void removeStatusItem(void) {
  if (statusItem != nil) {
    [[NSStatusBar systemStatusBar] removeStatusItem:statusItem];
    statusItem = nil;
    systrayMenu = nil;
    clickHandler = nil;
  }
}

void registerSystray(void) {
  if (statusItem != nil) {
    return;
  }
  dispatch_sync(dispatch_get_main_queue(), ^{
    statusItem = [[NSStatusBar systemStatusBar] statusItemWithLength:NSVariableStatusItemLength];
    systrayMenu = [[NSMenu alloc] init];
    [systrayMenu setAutoenablesItems:FALSE];
    [statusItem setMenu:systrayMenu];
    clickHandler = [[SystrayClickHandler alloc] init];
  });
  systray_ready();
}

int nativeLoop(void) {
  return 0;
}

void setIcon(const char* iconBytes, int length, bool template) {
  NSData* buffer = [NSData dataWithBytes:iconBytes length:length];
  NSImage *image = [[NSImage alloc] initWithData:buffer];
  [image setSize:NSMakeSize(16, 16)];
  image.template = template;
  dispatch_sync(dispatch_get_main_queue(), ^{
    statusItem.button.image = image;
    if (statusItem.button.title.length == 0) {
      statusItem.button.imagePosition = NSImageOnly;
    } else {
      statusItem.button.imagePosition = NSImageLeft;
    }
  });
}

void setMenuItemIcon(const char* iconBytes, int length, int menuId, bool template) {
  NSData* buffer = [NSData dataWithBytes:iconBytes length:length];
  NSImage *image = [[NSImage alloc] initWithData:buffer];
  [image setSize:NSMakeSize(16, 16)];
  image.template = template;
  NSNumber *mId = [NSNumber numberWithInt:menuId];
  dispatch_sync(dispatch_get_main_queue(), ^{
    NSMenuItem* item = find_menu_item(systrayMenu, mId);
    if (item != NULL) {
      item.image = image;
    }
  });
}

void setTitle(char* ctitle) {
  NSString* titleStr = [[NSString alloc] initWithCString:ctitle encoding:NSUTF8StringEncoding];
  free(ctitle);
  dispatch_sync(dispatch_get_main_queue(), ^{
    statusItem.button.title = titleStr;
    if (statusItem.button.image != nil) {
      if (statusItem.button.title.length == 0) {
        statusItem.button.imagePosition = NSImageOnly;
      } else {
        statusItem.button.imagePosition = NSImageLeft;
      }
    } else {
      statusItem.button.imagePosition = NSNoImage;
    }
  });
}

void setTooltip(char* ctooltip) {
  NSString* tooltipStr = [[NSString alloc] initWithCString:ctooltip encoding:NSUTF8StringEncoding];
  free(ctooltip);
  dispatch_sync(dispatch_get_main_queue(), ^{
    statusItem.button.toolTip = tooltipStr;
  });
}

void add_or_update_menu_item(int menuId, int parentMenuId, char* title, char* tooltip, short disabled, short checked, short isCheckable) {
  SystrayMenuItem* item = [[SystrayMenuItem alloc] initWithId:menuId withParentMenuId:parentMenuId withTitle:title withTooltip:tooltip withDisabled:disabled withChecked:checked];
  free(title);
  free(tooltip);
  dispatch_sync(dispatch_get_main_queue(), ^{
    NSMenu *theMenu = systrayMenu;
    if ([item->parentMenuId integerValue] > 0) {
      NSMenuItem *parentItem = find_menu_item(systrayMenu, item->parentMenuId);
      if (parentItem.hasSubmenu) {
        theMenu = parentItem.submenu;
      } else {
        theMenu = [[NSMenu alloc] init];
        [theMenu setAutoenablesItems:NO];
        [parentItem setSubmenu:theMenu];
      }
    }
    NSMenuItem *menuItem = find_menu_item(theMenu, item->menuId);
    if (menuItem == NULL) {
      menuItem = [theMenu addItemWithTitle:item->title
                                   action:@selector(menuHandler:)
                            keyEquivalent:@""];
      [menuItem setRepresentedObject:item->menuId];
      [menuItem setTarget:clickHandler];
    }
    [menuItem setTitle:item->title];
    [menuItem setTag:[item->menuId integerValue]];
    [menuItem setToolTip:item->tooltip];
    menuItem.enabled = (item->disabled != 1);
    menuItem.state = (item->checked == 1) ? NSControlStateValueOn : NSControlStateValueOff;
  });
}

void add_separator(int menuId) {
  dispatch_sync(dispatch_get_main_queue(), ^{
    [systrayMenu addItem:[NSMenuItem separatorItem]];
  });
}

void hide_menu_item(int menuId) {
  NSNumber *mId = [NSNumber numberWithInt:menuId];
  dispatch_sync(dispatch_get_main_queue(), ^{
    NSMenuItem* menuItem = find_menu_item(systrayMenu, mId);
    if (menuItem != NULL) {
      [menuItem setHidden:TRUE];
    }
  });
}

void show_menu_item(int menuId) {
  NSNumber *mId = [NSNumber numberWithInt:menuId];
  dispatch_sync(dispatch_get_main_queue(), ^{
    NSMenuItem* menuItem = find_menu_item(systrayMenu, mId);
    if (menuItem != NULL) {
      [menuItem setHidden:FALSE];
    }
  });
}

void stop_systray(void) {
  dispatch_sync(dispatch_get_main_queue(), ^{
    removeStatusItem();
  });
}

void forceQuit(void) {
  dispatch_async(dispatch_get_main_queue(), ^{
    [NSApp terminate:nil];
  });
}
