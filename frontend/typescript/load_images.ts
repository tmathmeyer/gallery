
import { GalleryImage, Gallery, SeedableRNG } from "./gallery";

export class RealImageLoader extends GalleryImage {

  GetTypeName() : string {
    return "ImageLoader";
  }

  GetDomElement() : HTMLElement {
    let _result = document.createElement('div');
    _result.style.background = "url("+this.imageUrl+")";
    _result.style.animation = 'fadein ' + (Math.random() * 7 + 1) + 's';
    _result.style['background-size'] = 'cover';
    _result.classList.add('GalleryImage-'+this.GetTypeName());
    return _result;
  }

  GetAspectRatio() : number {
    return 1;
  }
}

export class PreSizedFadeInImageLoader extends GalleryImage {

  constructor(url: string, private dimensions: [number, number]) {
    super(url);
  }

  GetTypeName(): string {
    return "PreSizedFadeInImageLoader"
  }

  GetDomElement(): HTMLElement {
    let _dom_element = document.createElement('div');
    _dom_element.style.animation = 'fadein 5s';
    _dom_element.style['background-size'] = 'cover';
    _dom_element.classList.add('GalleryImage-'+this.GetTypeName());

    let _image_preload = new Image();
    _image_preload.onload = function() {
      _dom_element.style['background-image'] = "url("+_image_preload.src+")";
    }
    _image_preload.src = this.imageUrl;

    return _dom_element;
  }

  GetAspectRatio(): number {
    return this.dimensions[0] / this.dimensions[1];
  }

}
